package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/toba/jig/internal/todo/integration/syncutil"
)

const baseURL = "https://api.github.com"

// Default retry configuration for rate limit handling
const (
	defaultMaxRetries     = 5
	defaultBaseRetryDelay = 1 * time.Second
	defaultMaxRetryDelay  = 30 * time.Second
)

// RateLimitError represents a GitHub rate limit error.
type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit: %s (retry after %v)", e.Message, e.RetryAfter)
}

// TransientError represents a transient error that can be retried.
type TransientError struct {
	Message string
}

func (e *TransientError) Error() string {
	return "transient error: " + e.Message
}

// RetryConfig holds retry settings for rate limit handling.
type RetryConfig struct {
	MaxRetries     int
	BaseRetryDelay time.Duration
	MaxRetryDelay  time.Duration
}

// Client provides GitHub API access via REST.
type Client struct {
	token      string
	owner      string
	repo       string
	httpClient *http.Client

	// Retry configuration (uses defaults if nil)
	retryConfig *RetryConfig

	// Cached authenticated user
	authenticatedUser *User
	// Cached labels (label name -> true)
	labelCache map[string]bool
}

func (c *Client) getRetryConfig() RetryConfig {
	if c.retryConfig != nil {
		return *c.retryConfig
	}
	return RetryConfig{
		MaxRetries:     defaultMaxRetries,
		BaseRetryDelay: defaultBaseRetryDelay,
		MaxRetryDelay:  defaultMaxRetryDelay,
	}
}

// newJSONRequest creates an HTTP request with a JSON-encoded body.
func (c *Client) newJSONRequest(ctx context.Context, method, url string, payload any) (*http.Request, error) {
	return syncutil.NewJSONRequest(ctx, method, url, payload)
}

// NewClient creates a new GitHub client.
func NewClient(token, owner, repo string) *Client {
	return &Client{
		token:      token,
		owner:      owner,
		repo:       repo,
		httpClient: &http.Client{},
	}
}

// GetRepo fetches repository metadata.
func (c *Client) GetRepo(ctx context.Context) (*Repo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", baseURL, c.owner, c.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp Repo
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting repo: %w", err)
	}

	return &resp, nil
}

// GetIssue fetches an issue by number.
func (c *Client) GetIssue(ctx context.Context, number int) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", baseURL, c.owner, c.repo, number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp Issue
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}

	return &resp, nil
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(ctx context.Context, issue *CreateIssueRequest) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, c.owner, c.repo)

	req, err := c.newJSONRequest(ctx, "POST", url, issue)
	if err != nil {
		return nil, err
	}

	var resp Issue
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("creating issue: %w", err)
	}

	return &resp, nil
}

// UpdateIssue updates an existing issue.
func (c *Client) UpdateIssue(ctx context.Context, number int, update *UpdateIssueRequest) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", baseURL, c.owner, c.repo, number)

	req, err := c.newJSONRequest(ctx, "PATCH", url, update)
	if err != nil {
		return nil, err
	}

	var resp Issue
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("updating issue: %w", err)
	}

	return &resp, nil
}

// GetAuthenticatedUser fetches the user associated with the API token.
// Results are cached for the lifetime of the client.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (*User, error) {
	if c.authenticatedUser != nil {
		return c.authenticatedUser, nil
	}

	url := baseURL + "/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp User
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting authenticated user: %w", err)
	}

	c.authenticatedUser = &resp
	return c.authenticatedUser, nil
}

// ListLabels fetches all labels for the repository.
func (c *Client) ListLabels(ctx context.Context) ([]Label, error) {
	var allLabels []Label
	page := 1

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/labels?per_page=100&page=%d", baseURL, c.owner, c.repo, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var labels []Label
		if err := c.doRequest(req, &labels); err != nil {
			return nil, fmt.Errorf("listing labels: %w", err)
		}

		allLabels = append(allLabels, labels...)
		if len(labels) < 100 {
			break
		}
		page++
	}

	return allLabels, nil
}

// CreateLabel creates a new label in the repository.
func (c *Client) CreateLabel(ctx context.Context, name, color string) (*Label, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/labels", baseURL, c.owner, c.repo)

	req, err := c.newJSONRequest(ctx, "POST", url, map[string]string{"name": name, "color": color})
	if err != nil {
		return nil, err
	}

	var resp Label
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("creating label: %w", err)
	}

	if c.labelCache != nil {
		c.labelCache[name] = true
	}

	return &resp, nil
}

// PopulateLabelCache fetches all labels and caches them.
func (c *Client) PopulateLabelCache(ctx context.Context) error {
	labels, err := c.ListLabels(ctx)
	if err != nil {
		return err
	}

	c.labelCache = make(map[string]bool, len(labels))
	for _, l := range labels {
		c.labelCache[l.Name] = true
	}

	return nil
}

// EnsureLabel creates a label if it doesn't exist in the cache.
func (c *Client) EnsureLabel(ctx context.Context, name, color string) error {
	if c.labelCache != nil && c.labelCache[name] {
		return nil
	}

	_, err := c.CreateLabel(ctx, name, color)
	if err != nil {
		// 422 means label already exists (race or cache miss)
		if strings.Contains(err.Error(), "422") || strings.Contains(err.Error(), "already_exists") {
			if c.labelCache == nil {
				c.labelCache = make(map[string]bool)
			}
			c.labelCache[name] = true
			return nil
		}
		return err
	}

	return nil
}

// AddSubIssue adds a sub-issue to a parent issue using the GitHub sub-issues API.
// If replaceParent is true, it will re-parent the child even if it already has a parent.
func (c *Client) AddSubIssue(ctx context.Context, parentNumber, childIssueID int, replaceParent bool) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/sub_issues", baseURL, c.owner, c.repo, parentNumber)

	body := &SubIssueRequest{SubIssueID: childIssueID, ReplaceParent: replaceParent}
	req, err := c.newJSONRequest(ctx, "POST", url, body)
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("adding sub-issue: %w", err)
	}

	return nil
}

// RemoveSubIssue removes a sub-issue from a parent issue.
func (c *Client) RemoveSubIssue(ctx context.Context, parentNumber, childIssueID int) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/sub_issue", baseURL, c.owner, c.repo, parentNumber)

	req, err := c.newJSONRequest(ctx, "DELETE", url, &RemoveSubIssueRequest{SubIssueID: childIssueID})
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("removing sub-issue: %w", err)
	}

	return nil
}

// GetParentIssue fetches the parent issue of a given issue, if any.
// Returns nil, nil if the issue has no parent.
func (c *Client) GetParentIssue(ctx context.Context, issueNumber int) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/parent", baseURL, c.owner, c.repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp Issue
	if err := c.doRequest(req, &resp); err != nil {
		// 404 means no parent
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			return nil, nil
		}
		return nil, fmt.Errorf("getting parent issue: %w", err)
	}

	return &resp, nil
}

// ListBlockedBy fetches the list of issues blocking the given issue.
func (c *Client) ListBlockedBy(ctx context.Context, issueNumber int) ([]BlockingDependency, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/dependencies/blocked_by", baseURL, c.owner, c.repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp []BlockingDependency
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("listing blocked-by: %w", err)
	}

	return resp, nil
}

// AddBlockedBy adds a blocked-by dependency to an issue.
func (c *Client) AddBlockedBy(ctx context.Context, issueNumber, blockerIssueID int) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/dependencies/blocked_by", baseURL, c.owner, c.repo, issueNumber)

	req, err := c.newJSONRequest(ctx, "POST", url, &AddBlockedByRequest{IssueID: blockerIssueID})
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("adding blocked-by: %w", err)
	}

	return nil
}

// RemoveBlockedBy removes a blocked-by dependency from an issue.
func (c *Client) RemoveBlockedBy(ctx context.Context, issueNumber, blockerIssueID int) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/dependencies/blocked_by/%d", baseURL, c.owner, c.repo, issueNumber, blockerIssueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("removing blocked-by: %w", err)
	}

	return nil
}

// ListBlocking fetches the list of issues that the given issue is blocking.
func (c *Client) ListBlocking(ctx context.Context, issueNumber int) ([]BlockingDependency, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/dependencies/blocking", baseURL, c.owner, c.repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp []BlockingDependency
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("listing blocking: %w", err)
	}

	return resp, nil
}

// CreateMilestone creates a new milestone.
func (c *Client) CreateMilestone(ctx context.Context, milestone *CreateMilestoneRequest) (*Milestone, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/milestones", baseURL, c.owner, c.repo)

	req, err := c.newJSONRequest(ctx, "POST", url, milestone)
	if err != nil {
		return nil, err
	}

	var resp Milestone
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("creating milestone: %w", err)
	}

	return &resp, nil
}

// UpdateMilestone updates an existing milestone.
func (c *Client) UpdateMilestone(ctx context.Context, number int, update *UpdateMilestoneRequest) (*Milestone, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/milestones/%d", baseURL, c.owner, c.repo, number)

	req, err := c.newJSONRequest(ctx, "PATCH", url, update)
	if err != nil {
		return nil, err
	}

	var resp Milestone
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("updating milestone: %w", err)
	}

	return &resp, nil
}

// GetMilestone fetches a milestone by number.
func (c *Client) GetMilestone(ctx context.Context, number int) (*Milestone, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/milestones/%d", baseURL, c.owner, c.repo, number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp Milestone
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting milestone: %w", err)
	}

	return &resp, nil
}

// ListMilestones fetches all milestones with the given state filter.
func (c *Client) ListMilestones(ctx context.Context, state string) ([]Milestone, error) {
	var allMilestones []Milestone
	page := 1

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/milestones?state=%s&per_page=100&page=%d", baseURL, c.owner, c.repo, state, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var milestones []Milestone
		if err := c.doRequest(req, &milestones); err != nil {
			return nil, fmt.Errorf("listing milestones: %w", err)
		}

		allMilestones = append(allMilestones, milestones...)
		if len(milestones) < 100 {
			break
		}
		page++
	}

	return allMilestones, nil
}

// doRequest executes an HTTP request and decodes the response.
// It delegates to syncutil.DoWithRetry with GitHub-specific auth and error handling.
func (c *Client) doRequest(req *http.Request, result any) error {
	cfg := c.getRetryConfig()

	hooks := syncutil.RequestHooks{
		SetAuth: func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+c.token)
			r.Header.Set("Accept", "application/vnd.github+json")
			r.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		},
		HandleRateLimit: func(resp *http.Response, body []byte) error {
			if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {
				retryAfter := 60 * time.Second // default
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if seconds, err := strconv.Atoi(ra); err == nil {
						retryAfter = time.Duration(seconds) * time.Second
					}
				}
				return &RateLimitError{
					Message:    string(body),
					RetryAfter: retryAfter,
				}
			}
			return nil
		},
		HandleAPIError: func(statusCode int, body []byte) error {
			var errResp errorResponse
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
				return fmt.Errorf("API error (HTTP %d): %s", statusCode, errResp.Message)
			}
			return nil
		},
	}

	return syncutil.DoWithRetry(c.httpClient, req, syncutil.RetryConfig{
		MaxRetries:     cfg.MaxRetries,
		BaseRetryDelay: cfg.BaseRetryDelay,
		MaxRetryDelay:  cfg.MaxRetryDelay,
	}, hooks, result)
}
