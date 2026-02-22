package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/toba/jig/internal/todo/integration/syncutil"
	"github.com/toba/jig/internal/todo/issue"
)

// memorySyncProvider is a simple in-memory SyncStateProvider for tests.
type memorySyncProvider struct {
	mu           sync.RWMutex
	issueNumbers map[string]int
	syncedAt     map[string]*time.Time
}

func newMemorySyncProvider() *memorySyncProvider {
	return &memorySyncProvider{
		issueNumbers: make(map[string]int),
		syncedAt:     make(map[string]*time.Time),
	}
}

func (m *memorySyncProvider) GetIssueNumber(issueID string) *int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.issueNumbers[issueID]
	if !ok || n == 0 {
		return nil
	}
	return &n
}

func (m *memorySyncProvider) GetSyncedAt(issueID string) *time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.syncedAt[issueID]
}

func (m *memorySyncProvider) SetIssueNumber(issueID string, number int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.issueNumbers[issueID] = number
}

func (m *memorySyncProvider) SetSyncedAt(issueID string, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	utc := t.UTC()
	m.syncedAt[issueID] = &utc
}

func (m *memorySyncProvider) Clear(issueID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.issueNumbers, issueID)
	delete(m.syncedAt, issueID)
}

func (m *memorySyncProvider) Flush() error { return nil }

func newTestSyncer(t *testing.T, client *Client) *Syncer {
	t.Helper()
	return &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{},
		syncStore:       newMemorySyncProvider(),
		issueToGHNumber: make(map[string]int),
		issueToGHID:     make(map[string]int),
		childrenOf:      make(map[string][]string),
	}
}

func TestComputeLabels(t *testing.T) {
	syncer := &Syncer{
		config: &Config{},
	}

	tests := []struct {
		name       string
		issue      *issue.Issue
		wantLabels []string
	}{
		{
			name: "only tags become labels",
			issue: &issue.Issue{
				Status:   "ready",
				Priority: "high",
				Type:     "bug",
				Tags:     []string{"urgent"},
			},
			wantLabels: []string{"urgent"},
		},
		{
			name: "no tags means no labels",
			issue: &issue.Issue{
				Status: "completed",
			},
			wantLabels: nil,
		},
		{
			name: "multiple tags",
			issue: &issue.Issue{
				Status: "unknown",
				Tags:   []string{"a", "b"},
			},
			wantLabels: []string{"a", "b"},
		},
		{
			name: "status and priority do not produce labels",
			issue: &issue.Issue{
				Status:   "draft",
				Priority: "critical",
			},
			wantLabels: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := syncer.computeLabels(tt.issue)
			if !slices.Equal(got, tt.wantLabels) {
				t.Errorf("computeLabels() = %v, want %v", got, tt.wantLabels)
			}
		})
	}
}


func TestBuildIssueBody(t *testing.T) {
	syncer := &Syncer{config: &Config{}}

	tests := []struct {
		name     string
		issue    *issue.Issue
		wantBody string
	}{
		{
			name:     "with body",
			issue:    &issue.Issue{ID: "test-1", Body: "Some description"},
			wantBody: "Some description\n\n" + syncutil.SyncFooter + "\n\n<!-- todo:test-1 -->",
		},
		{
			name:     "empty body",
			issue:    &issue.Issue{ID: "test-2"},
			wantBody: syncutil.SyncFooter + "\n\n<!-- todo:test-2 -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := syncer.buildIssueBody(tt.issue)
			if got != tt.wantBody {
				t.Errorf("buildIssueBody() = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestGetGitHubState(t *testing.T) {
	syncer := &Syncer{
		config: &Config{},
	}

	tests := []struct {
		status string
		want   string
	}{
		{"ready", "open"},
		{"draft", "open"},
		{"in-progress", "open"},
		{"completed", "closed"},
		{"scrapped", "closed"},
		{"unknown", "open"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := syncer.getGitHubState(tt.status)
			if got != tt.want {
				t.Errorf("getGitHubState(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestSyncIssue_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/issues") {
			resp := Issue{
				Number:  42,
				Title:   "Test",
				State:   "open",
				HTMLURL: "https://github.com/test/repo/issues/42",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/user" {
			resp := User{Login: "testuser", ID: 1}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test",
		owner: "test-owner",
		repo:  "test-repo",
		httpClient: &http.Client{
			Transport: &redirectTransport{target: server.URL},
		},
	}

	syncer := newTestSyncer(t, client)

	now := time.Now()
	b := &issue.Issue{
		ID:        "test-1",
		Title:     "Test issue",
		Status:    "ready",
		Type:      "task",
		Tags:      []string{"frontend"},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncIssue(context.Background(), b)

	if result.Action != "created" {
		t.Fatalf("expected action 'created', got %q", result.Action)
	}
	if result.ExternalID != "42" {
		t.Errorf("expected external ID '42', got %q", result.ExternalID)
	}
	if result.ExternalURL != "https://github.com/test/repo/issues/42" {
		t.Errorf("expected external URL, got %q", result.ExternalURL)
	}
}

func TestSyncIssue_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/issues/") {
			resp := Issue{
				Number:  42,
				Title:   "Old title",
				Body:    "old body\n\n<!-- todo:test-1 -->",
				State:   "open",
				HTMLURL: "https://github.com/test/repo/issues/42",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/issues/") {
			resp := Issue{
				Number:  42,
				Title:   "Updated issue",
				State:   "open",
				HTMLURL: "https://github.com/test/repo/issues/42",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test",
		owner: "test-owner",
		repo:  "test-repo",
		httpClient: &http.Client{
			Transport: &redirectTransport{target: server.URL},
		},
	}

	store := newMemorySyncProvider()
	store.SetIssueNumber("test-1", 42)
	syncer := &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{Force: true},
		syncStore:       store,
		issueToGHNumber: make(map[string]int),
		issueToGHID:     make(map[string]int),
	}

	now := time.Now()
	b := &issue.Issue{
		ID:        "test-1",
		Title:     "Updated issue",
		Status:    "ready",
		Type:      "task",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncIssue(context.Background(), b)

	if result.Action != "updated" {
		t.Fatalf("expected action 'updated', got %q", result.Action)
	}
}

func TestSyncIssue_CreateWithLabels(t *testing.T) {
	var receivedLabels []string
	var receivedType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/issues") {
			var req CreateIssueRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedLabels = req.Labels
			receivedType = req.Type

			resp := Issue{
				Number:  1,
				Title:   req.Title,
				State:   "open",
				HTMLURL: "https://github.com/test/repo/issues/1",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.URL.Path == "/user" {
			resp := User{Login: "testuser", ID: 1}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test",
		owner: "test-owner",
		repo:  "test-repo",
		httpClient: &http.Client{
			Transport: &redirectTransport{target: server.URL},
		},
	}

	syncer := &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{},
		syncStore:       newMemorySyncProvider(),
		issueToGHNumber: make(map[string]int),
		issueToGHID:     make(map[string]int),
	}

	now := time.Now()
	b := &issue.Issue{
		ID:        "test-1",
		Title:     "Test bug",
		Status:    "ready",
		Priority:  "high",
		Type:      "bug",
		Tags:      []string{"frontend"},
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncIssue(context.Background(), b)

	if result.Action != "created" {
		t.Fatalf("expected action 'created', got %q", result.Action)
	}

	// Only tags should appear as labels
	expected := []string{"frontend"}
	if !slices.Equal(receivedLabels, expected) {
		t.Errorf("labels = %v, want %v", receivedLabels, expected)
	}

	// Type should use GitHub's native type field
	if receivedType != "Bug" {
		t.Errorf("type = %q, want %q", receivedType, "Bug")
	}
}

func TestFilterIssuesNeedingSync(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	store := newMemorySyncProvider()
	store.SetIssueNumber("test-synced", 1)
	store.SetSyncedAt("test-synced", now)
	store.SetIssueNumber("test-stale", 2)
	store.SetSyncedAt("test-stale", past)

	issues := []*issue.Issue{
		{ID: "test-new", UpdatedAt: &now},
		{ID: "test-synced", UpdatedAt: &past},
		{ID: "test-stale", UpdatedAt: &future},
	}

	result := syncutil.FilterIssuesNeedingSync(issues, store, false)

	var ids []string
	for _, b := range result {
		ids = append(ids, b.ID)
	}
	sort.Strings(ids)

	expected := []string{"test-new", "test-stale"}
	sort.Strings(expected)

	if !slices.Equal(ids, expected) {
		t.Errorf("syncutil.FilterIssuesNeedingSync() = %v, want %v", ids, expected)
	}
}

func TestFilterIssuesNeedingSync_Force(t *testing.T) {
	now := time.Now()

	store := newMemorySyncProvider()
	store.SetIssueNumber("test-1", 1)
	store.SetSyncedAt("test-1", now)

	issues := []*issue.Issue{
		{ID: "test-1", UpdatedAt: &now},
		{ID: "test-2", UpdatedAt: &now},
	}

	result := syncutil.FilterIssuesNeedingSync(issues, store, true)

	if len(result) != 2 {
		t.Errorf("expected 2 issues with force=true, got %d", len(result))
	}
}

func TestSyncIssue_DryRun_Create(t *testing.T) {
	syncer := &Syncer{
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{DryRun: true},
		syncStore:       newMemorySyncProvider(),
		issueToGHNumber: make(map[string]int),
		issueToGHID:     make(map[string]int),
	}

	b := &issue.Issue{
		ID:     "test-1",
		Title:  "Test issue",
		Status: "ready",
	}

	result := syncer.syncIssue(context.Background(), b)

	if result.Action != "would create" {
		t.Fatalf("expected action 'would create', got %q", result.Action)
	}
}

func TestGetGitHubType(t *testing.T) {
	syncer := &Syncer{config: &Config{}}

	tests := []struct {
		issueType string
		want      string
	}{
		{"bug", "Bug"},
		{"feature", "Feature"},
		{"task", "Task"},
		{"milestone", "Task"},
		{"epic", "Task"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.issueType, func(t *testing.T) {
			got := syncer.getGitHubType(tt.issueType)
			if got != tt.want {
				t.Errorf("getGitHubType(%q) = %q, want %q", tt.issueType, got, tt.want)
			}
		})
	}
}

func TestBuildUpdateRequest_TypeChange(t *testing.T) {
	syncer := &Syncer{config: &Config{}}

	tests := []struct {
		name        string
		currentType *IssueType
		newType     string
		wantType    *string
	}{
		{
			name:        "type changed",
			currentType: &IssueType{Name: "Task"},
			newType:     "Bug",
			wantType:    new("Bug"),
		},
		{
			name:        "type unchanged",
			currentType: &IssueType{Name: "Bug"},
			newType:     "Bug",
			wantType:    nil,
		},
		{
			name:        "no current type, setting new",
			currentType: nil,
			newType:     "Feature",
			wantType:    new("Feature"),
		},
		{
			name:        "no current type, no new type",
			currentType: nil,
			newType:     "",
			wantType:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := &Issue{
				Title: "Test",
				Body:  "body\n\n<!-- todo:test-1 -->",
				State: "open",
				Type:  tt.currentType,
			}
			b := &issue.Issue{
				ID:    "test-1",
				Title: "Test",
			}
			update := syncer.buildUpdateRequest(current, b, "body\n\n<!-- todo:test-1 -->", "open", tt.newType, nil)
			if tt.wantType == nil && update.Type != nil {
				t.Errorf("expected nil Type, got %q", *update.Type)
			}
			if tt.wantType != nil {
				if update.Type == nil {
					t.Errorf("expected Type %q, got nil", *tt.wantType)
				} else if *update.Type != *tt.wantType {
					t.Errorf("Type = %q, want %q", *update.Type, *tt.wantType)
				}
			}
		})
	}
}
func TestStripRelationshipLines(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "strips all relationship lines",
			body: "Description\n\n**Parent:** #1\n**Children:** #2, #3\n**Blocks:** #4\n**Blocked by:** #5\n\n<!-- todo:test -->",
			want: "Description\n\n\n<!-- todo:test -->",
		},
		{
			name: "no relationship lines",
			body: "Just a description\n\n<!-- todo:test -->",
			want: "Just a description\n\n<!-- todo:test -->",
		},
		{
			name: "only blocks line",
			body: "Body\n\n**Blocks:** #10, #20\n\n<!-- todo:test -->",
			want: "Body\n\n\n<!-- todo:test -->",
		},
		{
			name: "empty body",
			body: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripRelationshipLines(tt.body)
			if got != tt.want {
				t.Errorf("stripRelationshipLines() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSyncRelationships_AllTypes(t *testing.T) {
	var updatedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/issues/") {
			resp := Issue{
				Number:  10,
				Body:    "Description\n\n" + syncutil.SyncFooter + "\n\n<!-- todo:issue-a -->",
				HTMLURL: "https://github.com/test/repo/issues/10",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/issues/") {
			var req UpdateIssueRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Body != nil {
				updatedBody = *req.Body
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Issue{Number: 10})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := &Syncer{
		client: client,
		config: &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:   SyncOptions{},
		issueToGHNumber: map[string]int{
			"issue-a":  10,
			"issue-b":  20,
			"issue-c":  30,
			"issue-d":  40,
			"parent-1": 50,
		},
		issueToGHID: make(map[string]int),
		childrenOf: map[string][]string{
			"issue-a": {"issue-c", "issue-d"},
		},
	}

	b := &issue.Issue{
		ID:        "issue-a",
		Parent:    "parent-1",
		Blocking:  []string{"issue-b"},
		BlockedBy: []string{"issue-c"},
	}

	err := syncer.syncRelationships(context.Background(), b)
	if err != nil {
		t.Fatalf("syncRelationships() error: %v", err)
	}

	// Check that all relationship lines are present
	if !strings.Contains(updatedBody, "**Parent:** #50") {
		t.Errorf("missing Parent line in body: %s", updatedBody)
	}
	if !strings.Contains(updatedBody, "**Children:** #30, #40") {
		t.Errorf("missing Children line in body: %s", updatedBody)
	}
	if !strings.Contains(updatedBody, "**Blocks:** #20") {
		t.Errorf("missing Blocks line in body: %s", updatedBody)
	}
	if !strings.Contains(updatedBody, "**Blocked by:** #30") {
		t.Errorf("missing Blocked-by line in body: %s", updatedBody)
	}

	// Relationship lines should come before the todo comment
	todoIdx := strings.Index(updatedBody, "<!-- todo:")
	blocksIdx := strings.Index(updatedBody, "**Blocks:**")
	if blocksIdx > todoIdx {
		t.Error("relationship lines should come before <!-- todo: --> comment")
	}
}

func TestSyncRelationships_CleansStaleLines(t *testing.T) {
	var updatedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/issues/") {
			// Body has old relationship lines that should be removed
			resp := Issue{
				Number: 10,
				Body:   "Description\n\n**Blocks:** #99\n**Blocked by:** #88\n\n" + syncutil.SyncFooter + "\n\n<!-- todo:issue-a -->",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/issues/") {
			var req UpdateIssueRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Body != nil {
				updatedBody = *req.Body
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Issue{Number: 10})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{},
		issueToGHNumber: map[string]int{"issue-a": 10},
		issueToGHID:     make(map[string]int),
		childrenOf:      make(map[string][]string),
	}

	// Issue with no relationships (all removed)
	b := &issue.Issue{ID: "issue-a"}

	err := syncer.syncRelationships(context.Background(), b)
	if err != nil {
		t.Fatalf("syncRelationships() error: %v", err)
	}

	// Old relationship lines should be removed
	if strings.Contains(updatedBody, "**Blocks:**") {
		t.Errorf("stale Blocks line should be removed: %s", updatedBody)
	}
	if strings.Contains(updatedBody, "**Blocked by:**") {
		t.Errorf("stale Blocked-by line should be removed: %s", updatedBody)
	}
	// Description and footer should remain
	if !strings.Contains(updatedBody, "Description") {
		t.Errorf("description should be preserved: %s", updatedBody)
	}
}

func TestSyncSubIssueLink_AddParent(t *testing.T) {
	var addedSubIssue bool
	var addedParentNumber int
	var addedChildIssueID int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GetParentIssue returns 404 (no current parent)
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/parent") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
			return
		}
		// AddSubIssue
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/sub_issues") {
			addedSubIssue = true
			// Extract parent number from URL
			parts := strings.Split(r.URL.Path, "/")
			for i, p := range parts {
				if p == "issues" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &addedParentNumber)
					break
				}
			}
			// Extract sub_issue_id from request body
			var body SubIssueRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			addedChildIssueID = body.SubIssueID
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Issue{Number: 50})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{},
		issueToGHNumber: map[string]int{"parent-1": 50, "child-1": 10},
		issueToGHID:     map[string]int{"child-1": 100010},
		childrenOf:      make(map[string][]string),
	}

	b := &issue.Issue{ID: "child-1", Parent: "parent-1"}
	syncer.syncSubIssueLink(context.Background(), b, 10)

	if !addedSubIssue {
		t.Error("expected AddSubIssue to be called")
	}
	if addedParentNumber != 50 {
		t.Errorf("expected parent number 50, got %d", addedParentNumber)
	}
	if addedChildIssueID != 100010 {
		t.Errorf("expected child issue ID 100010 (not number 10), got %d", addedChildIssueID)
	}
}

func TestSyncSubIssueLink_RemoveParent(t *testing.T) {
	var removedSubIssue bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GetParentIssue returns a parent (issue has one on GitHub)
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/parent") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Issue{Number: 50})
			return
		}
		// RemoveSubIssue
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/sub_issue") {
			removedSubIssue = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Issue{Number: 50})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	client := &Client{
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := &Syncer{
		client:          client,
		config:          &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:            SyncOptions{},
		issueToGHNumber: map[string]int{"child-1": 10},
		issueToGHID:     map[string]int{"child-1": 100010},
		childrenOf:      make(map[string][]string),
	}

	// Issue with no parent locally (removed)
	b := &issue.Issue{ID: "child-1"}
	syncer.syncSubIssueLink(context.Background(), b, 10)

	if !removedSubIssue {
		t.Error("expected RemoveSubIssue to be called")
	}
}

func TestSyncSubIssueLink_NoRelationships(t *testing.T) {
	syncer := &Syncer{
		opts:            SyncOptions{NoRelationships: true},
		issueToGHNumber: map[string]int{"child-1": 10, "parent-1": 50},
		issueToGHID:     map[string]int{"child-1": 100010},
	}

	// Should return immediately without making any API calls
	b := &issue.Issue{ID: "child-1", Parent: "parent-1"}
	syncer.syncSubIssueLink(context.Background(), b, 10)
	// If it tried to call the client, it would panic (client is nil)
}

// redirectTransport redirects all requests to the test server.
type redirectTransport struct {
	target string
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(rt.target, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
