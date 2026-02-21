package clickup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/toba/jig/internal/todo/integration/syncutil"
)

const baseURL = "https://api.clickup.com/api/v2"

// Default retry configuration for rate limit handling
const (
	defaultMaxRetries     = 5
	defaultBaseRetryDelay = 1 * time.Second
	defaultMaxRetryDelay  = 30 * time.Second
)

// RateLimitError represents a ClickUp rate limit error.
type RateLimitError struct {
	Message string
	Code    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit: %s (code: %s)", e.Message, e.Code)
}

// TransientError represents a transient error that can be retried.
type TransientError struct {
	Message string
}

func (e *TransientError) Error() string {
	return fmt.Sprintf("transient error: %s", e.Message)
}

// RetryConfig holds retry settings for rate limit handling.
type RetryConfig struct {
	MaxRetries     int
	BaseRetryDelay time.Duration
	MaxRetryDelay  time.Duration
}

// Client provides ClickUp API access via REST.
type Client struct {
	token      string
	httpClient *http.Client

	// Retry configuration (uses defaults if nil)
	retryConfig *RetryConfig

	// Cached list info
	listInfo *List
	// Cached authorized user
	authorizedUser *AuthorizedUser
	// Cached space tags (tag name -> true)
	spaceTags map[string]bool
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

// NewClient creates a new ClickUp client.
// The token should be a ClickUp API token.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// GetList fetches list metadata including available statuses.
func (c *Client) GetList(ctx context.Context, listID string) (*List, error) {
	if c.listInfo != nil && c.listInfo.ID == listID {
		return c.listInfo, nil
	}

	url := fmt.Sprintf("%s/list/%s", baseURL, listID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp listResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting list: %w", err)
	}

	c.listInfo = &List{
		ID:       resp.ID,
		Name:     resp.Name,
		SpaceID:  resp.Space.ID,
		Statuses: resp.Statuses,
	}

	return c.listInfo, nil
}

// GetTask fetches a task by ID.
func (c *Client) GetTask(ctx context.Context, taskID string) (*TaskInfo, error) {
	url := fmt.Sprintf("%s/task/%s", baseURL, taskID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp taskResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}

	return resp.toTaskInfo(), nil
}

// CreateTask creates a new task in the given list.
func (c *Client) CreateTask(ctx context.Context, listID string, task *CreateTaskRequest) (*TaskInfo, error) {
	url := fmt.Sprintf("%s/list/%s/task", baseURL, listID)

	req, err := c.newJSONRequest(ctx, "POST", url, task)
	if err != nil {
		return nil, err
	}

	var resp taskResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	return resp.toTaskInfo(), nil
}

// UpdateTask updates an existing task.
func (c *Client) UpdateTask(ctx context.Context, taskID string, update *UpdateTaskRequest) (*TaskInfo, error) {
	url := fmt.Sprintf("%s/task/%s", baseURL, taskID)

	req, err := c.newJSONRequest(ctx, "PUT", url, update)
	if err != nil {
		return nil, err
	}

	var resp taskResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}

	return resp.toTaskInfo(), nil
}

// AddDependency adds a dependency to a task.
// This sets the task with taskID as waiting on (depends on) the task with dependsOnID.
// In other words: dependsOnID is blocking taskID.
func (c *Client) AddDependency(ctx context.Context, taskID, dependsOnID string) error {
	url := fmt.Sprintf("%s/task/%s/dependency", baseURL, taskID)

	req, err := c.newJSONRequest(ctx, "POST", url, &AddDependencyRequest{DependsOn: dependsOnID})
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("adding dependency: %w", err)
	}

	return nil
}

// GetAuthorizedUser fetches the user associated with the API token.
// Results are cached for the lifetime of the client.
func (c *Client) GetAuthorizedUser(ctx context.Context) (*AuthorizedUser, error) {
	if c.authorizedUser != nil {
		return c.authorizedUser, nil
	}

	url := fmt.Sprintf("%s/user", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp userResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting authorized user: %w", err)
	}

	c.authorizedUser = &resp.User
	return c.authorizedUser, nil
}

// GetAccessibleCustomFields fetches available custom fields for a list.
func (c *Client) GetAccessibleCustomFields(ctx context.Context, listID string) ([]FieldInfo, error) {
	url := fmt.Sprintf("%s/list/%s/field", baseURL, listID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp fieldsResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting custom fields: %w", err)
	}

	return resp.Fields, nil
}

// GetCustomItems fetches custom task types from all accessible workspaces.
// Returns custom items with their IDs, names, and descriptions.
func (c *Client) GetCustomItems(ctx context.Context) ([]CustomItem, error) {
	// First get all teams to iterate through workspaces
	teamsURL := fmt.Sprintf("%s/team", baseURL)
	teamsReq, err := http.NewRequestWithContext(ctx, "GET", teamsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating teams request: %w", err)
	}

	var teamsResp teamsResponse
	if err := c.doRequest(teamsReq, &teamsResp); err != nil {
		return nil, fmt.Errorf("getting teams: %w", err)
	}

	// Collect custom items from all teams
	seen := make(map[int]bool)
	var items []CustomItem
	for _, team := range teamsResp.Teams {
		url := fmt.Sprintf("%s/team/%s/custom_item", baseURL, team.ID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		var resp customItemsResponse
		if err := c.doRequest(req, &resp); err != nil {
			// Skip teams that don't support custom items
			continue
		}

		for _, item := range resp.CustomItems {
			if !seen[item.ID] {
				seen[item.ID] = true
				items = append(items, item)
			}
		}
	}

	return items, nil
}

// AddTagToTask adds a tag to a task.
// Note: This creates a task-level tag but does NOT register it as a space-level tag.
// Use EnsureSpaceTag before this to make tags discoverable in the space tag picker.
func (c *Client) AddTagToTask(ctx context.Context, taskID, tagName string) error {
	url := fmt.Sprintf("%s/task/%s/tag/%s", baseURL, taskID, tagName)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("adding tag: %w", err)
	}

	return nil
}

// RemoveTagFromTask removes a tag from a task.
func (c *Client) RemoveTagFromTask(ctx context.Context, taskID, tagName string) error {
	url := fmt.Sprintf("%s/task/%s/tag/%s", baseURL, taskID, tagName)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("removing tag: %w", err)
	}

	return nil
}

// GetSpaceTags fetches all tags for a space.
func (c *Client) GetSpaceTags(ctx context.Context, spaceID string) ([]Tag, error) {
	url := fmt.Sprintf("%s/space/%s/tag", baseURL, spaceID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	var resp spaceTagsResponse
	if err := c.doRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("getting space tags: %w", err)
	}

	return resp.Tags, nil
}

// CreateSpaceTag creates a tag at the space level so it appears in the tag picker.
func (c *Client) CreateSpaceTag(ctx context.Context, spaceID, tagName string) error {
	url := fmt.Sprintf("%s/space/%s/tag", baseURL, spaceID)

	req, err := c.newJSONRequest(ctx, "POST", url, map[string]any{"tag": map[string]string{"name": tagName}})
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("creating space tag: %w", err)
	}

	return nil
}

// PopulateSpaceTagCache fetches existing space tags into the client cache.
func (c *Client) PopulateSpaceTagCache(ctx context.Context, spaceID string) error {
	tags, err := c.GetSpaceTags(ctx, spaceID)
	if err != nil {
		return err
	}

	c.spaceTags = make(map[string]bool, len(tags))
	for _, t := range tags {
		c.spaceTags[t.Name] = true
	}

	return nil
}

// EnsureSpaceTag creates a tag at the space level if it doesn't already exist in the cache.
func (c *Client) EnsureSpaceTag(ctx context.Context, spaceID, tagName string) error {
	if c.spaceTags != nil && c.spaceTags[tagName] {
		return nil
	}

	if err := c.CreateSpaceTag(ctx, spaceID, tagName); err != nil {
		return err
	}

	if c.spaceTags == nil {
		c.spaceTags = make(map[string]bool)
	}
	c.spaceTags[tagName] = true

	return nil
}

// HasSpaceTag returns true if the tag exists in the space tag cache.
// PopulateSpaceTagCache must be called first.
func (c *Client) HasSpaceTag(tagName string) bool {
	return c.spaceTags != nil && c.spaceTags[tagName]
}

// SetCustomFieldValue sets a custom field value on a task.
func (c *Client) SetCustomFieldValue(ctx context.Context, taskID, fieldID string, value any) error {
	url := fmt.Sprintf("%s/task/%s/field/%s", baseURL, taskID, fieldID)

	req, err := c.newJSONRequest(ctx, "POST", url, map[string]any{"value": value})
	if err != nil {
		return err
	}

	if err := c.doRequest(req, nil); err != nil {
		return fmt.Errorf("setting custom field: %w", err)
	}

	return nil
}

// doRequest executes an HTTP request and decodes the response.
// It delegates to syncutil.DoWithRetry with ClickUp-specific auth and error handling.
func (c *Client) doRequest(req *http.Request, result any) error {
	cfg := c.getRetryConfig()

	hooks := syncutil.RequestHooks{
		SetAuth: func(r *http.Request) {
			r.Header.Set("Authorization", c.token)
		},
		HandleRateLimit: func(resp *http.Response, body []byte) error {
			var errResp errorResponse
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Err != "" {
				if resp.StatusCode == 429 || errResp.ECODE == "APP_002" {
					return &RateLimitError{Message: errResp.Err, Code: errResp.ECODE}
				}
			}
			return nil
		},
		HandleAPIError: func(statusCode int, body []byte) error {
			var errResp errorResponse
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Err != "" {
				return fmt.Errorf("API error: %s (code: %s)", errResp.Err, errResp.ECODE)
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
