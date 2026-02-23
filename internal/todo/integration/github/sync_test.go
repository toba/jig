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
	mu               sync.RWMutex
	issueNumbers     map[string]int
	milestoneNumbers map[string]int
	syncedAt         map[string]*time.Time
}

func newMemorySyncProvider() *memorySyncProvider {
	return &memorySyncProvider{
		issueNumbers:     make(map[string]int),
		milestoneNumbers: make(map[string]int),
		syncedAt:         make(map[string]*time.Time),
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

func (m *memorySyncProvider) GetMilestoneNumber(issueID string) *int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.milestoneNumbers[issueID]
	if !ok || n == 0 {
		return nil
	}
	return &n
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

func (m *memorySyncProvider) SetMilestoneNumber(issueID string, number int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.milestoneNumbers[issueID] = number
}

func (m *memorySyncProvider) Clear(issueID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.issueNumbers, issueID)
	delete(m.milestoneNumbers, issueID)
	delete(m.syncedAt, issueID)
}

func (m *memorySyncProvider) Flush() error { return nil }

func newTestSyncer(t *testing.T, client *Client) *Syncer {
	t.Helper()
	return &Syncer{
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		syncStore:              newMemorySyncProvider(),
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		childrenOf:             make(map[string][]string),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{Force: true},
		syncStore:              store,
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		syncStore:              newMemorySyncProvider(),
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{DryRun: true},
		syncStore:              newMemorySyncProvider(),
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
		{"epic", "Task"},
		{"milestone", ""}, // milestone-type issues don't map to GitHub issue types
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
	syncer := &Syncer{config: &Config{}, issueTypes: make(map[string]string)}

	strPtr := func(s string) *string { return &s }

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
			wantType:    strPtr("Bug"),
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
			wantType:    strPtr("Feature"),
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
			update := syncer.buildUpdateRequest(current, b, "body\n\n<!-- todo:test-1 -->", "open", tt.newType, nil, nil)
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

func TestBuildUpdateRequest_MilestoneClear(t *testing.T) {
	syncer := &Syncer{config: &Config{}, issueTypes: make(map[string]string)}

	// GitHub issue has a milestone, but local issue has no milestone parent
	current := &Issue{
		Title:     "Test",
		Body:      "body\n\n<!-- todo:test-1 -->",
		State:     "open",
		Milestone: &Milestone{Number: 1},
	}
	b := &issue.Issue{
		ID:    "test-1",
		Title: "Test",
	}
	update := syncer.buildUpdateRequest(current, b, "body\n\n<!-- todo:test-1 -->", "open", "", nil, nil)

	// Milestone should be set (to clear it)
	if !update.Milestone.Set {
		t.Fatal("expected Milestone.Set to be true")
	}
	if update.Milestone.Value != 0 {
		t.Errorf("expected Milestone.Value = 0, got %d", update.Milestone.Value)
	}

	// Verify JSON serialization produces null, not 0
	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	dataStr := string(data)
	if !strings.Contains(dataStr, `"milestone":null`) {
		t.Errorf("expected milestone:null in JSON, got %s", dataStr)
	}
	if strings.Contains(dataStr, `"milestone":0`) {
		t.Errorf("milestone must not be 0 in JSON (GitHub rejects it), got %s", dataStr)
	}
}

func TestBuildUpdateRequest_MilestoneChange(t *testing.T) {
	syncer := &Syncer{config: &Config{}, issueTypes: make(map[string]string)}

	milestoneNum := 2
	current := &Issue{
		Title:     "Test",
		Body:      "body\n\n<!-- todo:test-1 -->",
		State:     "open",
		Milestone: &Milestone{Number: 1},
	}
	b := &issue.Issue{
		ID:    "test-1",
		Title: "Test",
	}
	update := syncer.buildUpdateRequest(current, b, "body\n\n<!-- todo:test-1 -->", "open", "", nil, &milestoneNum)

	if !update.Milestone.Set {
		t.Fatal("expected Milestone.Set to be true")
	}
	if update.Milestone.Value != 2 {
		t.Errorf("expected Milestone.Value = 2, got %d", update.Milestone.Value)
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"milestone":2`) {
		t.Errorf("expected milestone:2 in JSON, got %s", string(data))
	}
}

func TestBuildUpdateRequest_MilestoneUnchanged(t *testing.T) {
	syncer := &Syncer{config: &Config{}, issueTypes: make(map[string]string)}

	milestoneNum := 1
	current := &Issue{
		Title:     "Test",
		Body:      "body\n\n<!-- todo:test-1 -->",
		State:     "open",
		Milestone: &Milestone{Number: 1},
	}
	b := &issue.Issue{
		ID:    "test-1",
		Title: "Test",
	}
	update := syncer.buildUpdateRequest(current, b, "body\n\n<!-- todo:test-1 -->", "open", "", nil, &milestoneNum)

	if update.Milestone.Set {
		t.Error("expected Milestone.Set to be false when milestone unchanged")
	}
	if update.hasChanges() {
		t.Error("expected no changes when nothing differs")
	}
}

func TestBuildUpdateRequest_StripsStaleFooterLinks(t *testing.T) {
	syncer := &Syncer{
		config:     &Config{},
		issueTypes: make(map[string]string),
	}

	// GitHub issue body has stale relationship lines
	current := &Issue{
		Title: "Test",
		Body:  "Description\n\n**Blocks:** #99\n**Blocked by:** #88\n\n" + syncutil.SyncFooter + "\n\n<!-- todo:test-1 -->",
		State: "open",
	}
	b := &issue.Issue{
		ID:    "test-1",
		Title: "Test",
		Body:  "Description",
	}
	// The desired body has no relationship lines
	desiredBody := "Description\n\n" + syncutil.SyncFooter + "\n\n<!-- todo:test-1 -->"

	update := syncer.buildUpdateRequest(current, b, desiredBody, "open", "", nil, nil)

	// Body should be updated because stale lines are stripped from current before comparison
	if update.Body == nil {
		t.Error("expected Body update to strip stale relationship lines")
	}
	if update.Body != nil && *update.Body != desiredBody {
		t.Errorf("Body = %q, want %q", *update.Body, desiredBody)
	}
}

func TestSyncMilestone_Create(t *testing.T) {
	var receivedReq CreateMilestoneRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/milestones") {
			_ = json.NewDecoder(r.Body).Decode(&receivedReq)
			resp := Milestone{
				ID:      1,
				Number:  5,
				Title:   receivedReq.Title,
				State:   "open",
				HTMLURL: "https://github.com/test/repo/milestone/5",
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
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := newTestSyncer(t, client)

	now := time.Now()
	b := &issue.Issue{
		ID:        "ms-1",
		Title:     "v1.0 Release",
		Status:    "ready",
		Type:      "milestone",
		Body:      "First major release",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncMilestone(context.Background(), b)

	if result.Action != "created" {
		t.Fatalf("expected action 'created', got %q", result.Action)
	}
	if result.ExternalID != "milestone:5" {
		t.Errorf("expected external ID 'milestone:5', got %q", result.ExternalID)
	}
	if receivedReq.Title != "v1.0 Release" {
		t.Errorf("expected title 'v1.0 Release', got %q", receivedReq.Title)
	}
	if !strings.Contains(receivedReq.Description, "First major release") {
		t.Errorf("expected description to contain body, got %q", receivedReq.Description)
	}
	// Verify milestone number stored
	if syncer.issueToMilestoneNumber["ms-1"] != 5 {
		t.Errorf("expected milestone number 5 in map, got %d", syncer.issueToMilestoneNumber["ms-1"])
	}
}

func TestSyncMilestone_Update(t *testing.T) {
	var patchCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/milestones/") {
			resp := Milestone{
				ID:          1,
				Number:      5,
				Title:       "Old Title",
				Description: "old desc\n\n<!-- todo:ms-1 -->",
				State:       "open",
				HTMLURL:     "https://github.com/test/repo/milestone/5",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/milestones/") {
			patchCalled = true
			resp := Milestone{
				Number:  5,
				Title:   "v1.0 Release",
				HTMLURL: "https://github.com/test/repo/milestone/5",
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
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	store := newMemorySyncProvider()
	store.SetMilestoneNumber("ms-1", 5)
	syncer := &Syncer{
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{Force: true},
		syncStore:              store,
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
	}

	now := time.Now()
	b := &issue.Issue{
		ID:        "ms-1",
		Title:     "v1.0 Release",
		Status:    "ready",
		Type:      "milestone",
		Body:      "Updated description",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncMilestone(context.Background(), b)

	if result.Action != "updated" {
		t.Fatalf("expected action 'updated', got %q", result.Action)
	}
	if !patchCalled {
		t.Error("expected PATCH to be called for milestone update")
	}
}

func TestSyncIssue_MilestoneAssignment(t *testing.T) {
	var receivedMilestone *int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/issues") {
			var req CreateIssueRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedMilestone = req.Milestone

			resp := Issue{
				Number:  10,
				Title:   req.Title,
				State:   "open",
				HTMLURL: "https://github.com/test/repo/issues/10",
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
		token: "test", owner: "test-owner", repo: "test-repo",
		httpClient: &http.Client{Transport: &redirectTransport{target: server.URL}},
	}

	syncer := &Syncer{
		client:    client,
		config:    &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:      SyncOptions{NoRelationships: true},
		syncStore: newMemorySyncProvider(),
		issueToGHNumber: make(map[string]int),
		issueToGHID:     make(map[string]int),
		issueToMilestoneNumber: map[string]int{
			"ms-1": 5, // Parent milestone has GitHub milestone number 5
		},
		issueTypes: map[string]string{
			"ms-1":   "milestone",
			"task-1": "task",
		},
	}

	now := time.Now()
	b := &issue.Issue{
		ID:        "task-1",
		Title:     "Implement feature X",
		Status:    "ready",
		Type:      "task",
		Parent:    "ms-1", // Parent is a milestone
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	result := syncer.syncIssue(context.Background(), b)

	if result.Action != "created" {
		t.Fatalf("expected action 'created', got %q", result.Action)
	}
	if receivedMilestone == nil {
		t.Fatal("expected milestone to be set on created issue")
	}
	if *receivedMilestone != 5 {
		t.Errorf("expected milestone number 5, got %d", *receivedMilestone)
	}
}

func TestSyncBlockingRelationships_Add(t *testing.T) {
	var addedBlockedByCalls []struct {
		issueNumber int
		blockerID   int
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ListBlockedBy returns empty (no current deps)
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/blocked_by") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]BlockingDependency{})
			return
		}
		// AddBlockedBy
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/blocked_by") {
			var req AddBlockedByRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			// Extract issue number from URL
			parts := strings.Split(r.URL.Path, "/")
			var num int
			for i, p := range parts {
				if p == "issues" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &num)
					break
				}
			}
			addedBlockedByCalls = append(addedBlockedByCalls, struct {
				issueNumber int
				blockerID   int
			}{num, req.IssueID})
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(BlockingDependency{ID: req.IssueID, Number: num})
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"issue-a": 10, "issue-b": 20},
		issueToGHID:            map[string]int{"issue-a": 100010, "issue-b": 100020},
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
	}

	// issue-a is blocked by issue-b
	b := &issue.Issue{
		ID:        "issue-a",
		BlockedBy: []string{"issue-b"},
	}

	syncer.syncBlockingRelationships(context.Background(), b)

	// Should have called AddBlockedBy on issue-a with blocker issue-b's GH ID
	found := false
	for _, call := range addedBlockedByCalls {
		if call.issueNumber == 10 && call.blockerID == 100020 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected AddBlockedBy(10, 100020) to be called, got calls: %+v", addedBlockedByCalls)
	}
}

func TestSyncBlockingRelationships_Remove(t *testing.T) {
	var removedCalls []struct {
		issueNumber int
		blockerID   int
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ListBlockedBy returns an existing dependency that should be removed
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/blocked_by") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]BlockingDependency{
				{ID: 100099, Number: 99}, // stale dependency
			})
			return
		}
		// RemoveBlockedBy
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/blocked_by/") {
			parts := strings.Split(r.URL.Path, "/")
			var issueNum, blockerID int
			for i, p := range parts {
				if p == "issues" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &issueNum)
				}
				if p == "blocked_by" && i+1 < len(parts) {
					fmt.Sscanf(parts[i+1], "%d", &blockerID)
				}
			}
			removedCalls = append(removedCalls, struct {
				issueNumber int
				blockerID   int
			}{issueNum, blockerID})
			w.WriteHeader(204)
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"issue-a": 10},
		issueToGHID:            map[string]int{"issue-a": 100010},
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
	}

	// issue-a has no blocking relationships (all removed)
	b := &issue.Issue{
		ID:        "issue-a",
		BlockedBy: []string{}, // empty — should remove stale dep
	}

	syncer.syncBlockingRelationships(context.Background(), b)

	if len(removedCalls) != 1 {
		t.Fatalf("expected 1 RemoveBlockedBy call, got %d", len(removedCalls))
	}
	if removedCalls[0].issueNumber != 10 || removedCalls[0].blockerID != 100099 {
		t.Errorf("expected RemoveBlockedBy(10, 100099), got (%d, %d)", removedCalls[0].issueNumber, removedCalls[0].blockerID)
	}
}

func TestSyncBlockingRelationships_NoChange(t *testing.T) {
	apiCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		// ListBlockedBy returns current dep that matches desired
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/blocked_by") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]BlockingDependency{
				{ID: 100020, Number: 20},
			})
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"issue-a": 10, "issue-b": 20},
		issueToGHID:            map[string]int{"issue-a": 100010, "issue-b": 100020},
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
	}

	// issue-a is blocked by issue-b (matches current state)
	b := &issue.Issue{
		ID:        "issue-a",
		BlockedBy: []string{"issue-b"},
	}

	syncer.syncBlockingRelationships(context.Background(), b)

	// Should only have the GET call, no POST or DELETE
	if apiCalls != 1 {
		t.Errorf("expected 1 API call (GET only), got %d", apiCalls)
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"parent-1": 50, "child-1": 10},
		issueToGHID:            map[string]int{"child-1": 100010},
		childrenOf:             make(map[string][]string),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             map[string]string{"parent-1": "feature"},
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

func TestSyncSubIssueLink_SkipsMilestoneParent(t *testing.T) {
	// If the client is called, it would panic (client is nil)
	syncer := &Syncer{
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"ms-1": 50, "child-1": 10},
		issueToGHID:            map[string]int{"child-1": 100010},
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             map[string]string{"ms-1": "milestone"},
	}

	// Child's parent is a milestone — should skip sub-issue linking
	b := &issue.Issue{ID: "child-1", Parent: "ms-1"}
	syncer.syncSubIssueLink(context.Background(), b, 10)
	// No panic means it correctly skipped
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
		client:                 client,
		config:                 &Config{Owner: "test-owner", Repo: "test-repo"},
		opts:                   SyncOptions{},
		issueToGHNumber:        map[string]int{"child-1": 10},
		issueToGHID:            map[string]int{"child-1": 100010},
		childrenOf:             make(map[string][]string),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
		opts:                   SyncOptions{NoRelationships: true},
		issueToGHNumber:        map[string]int{"child-1": 10, "parent-1": 50},
		issueToGHID:            map[string]int{"child-1": 100010},
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
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
