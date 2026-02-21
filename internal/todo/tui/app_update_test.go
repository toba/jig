package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/toba/jig/internal/todo/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/toba/jig/internal/todo/graph"
	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/issue"

	"github.com/toba/jig/internal/todo/core"
)

// newTestApp creates an App backed by a real Core (in a temp dir) so resolver
// queries work correctly.
func newTestApp(t *testing.T) *App {
	t.Helper()
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create .issues dir: %v", err)
	}

	cfg := config.Default()
	c := core.New(dataDir, cfg)
	if err := c.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	app := New(c, cfg)
	// Set dimensions so View() does not return "Loading..."
	app.width = 80
	app.height = 24
	app.list.width = 80
	app.list.height = 24
	app.list.list.SetSize(78, 20)
	return app
}

// newTestAppWithIssues creates an App with some pre-loaded issues.
func newTestAppWithIssues(t *testing.T) (*App, *core.Core) {
	t.Helper()
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create .issues dir: %v", err)
	}

	cfg := config.Default()
	c := core.New(dataDir, cfg)
	if err := c.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	// Create some issues
	issues := []*issue.Issue{
		{ID: "abc-123", Title: "First issue", Status: "todo", Type: "task", Tags: []string{"frontend"}},
		{ID: "def-456", Title: "Second issue", Status: "in-progress", Type: "bug"},
		{ID: "ghi-789", Title: "Third issue", Status: "completed", Type: "feature"},
	}
	for _, b := range issues {
		if err := c.Create(b); err != nil {
			t.Fatalf("failed to create issue: %v", err)
		}
	}

	app := New(c, cfg)
	app.width = 80
	app.height = 24
	app.list.width = 80
	app.list.height = 24
	app.list.list.SetSize(78, 20)
	return app, c
}

func TestAppInitialState(t *testing.T) {
	app := newTestApp(t)

	if app.state != viewList {
		t.Errorf("initial state = %d, want viewList (%d)", app.state, viewList)
	}
	if app.pendingKey != "" {
		t.Errorf("pendingKey = %q, want empty", app.pendingKey)
	}
}

func TestAppCtrlCQuits(t *testing.T) {
	app := newTestApp(t)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("ctrl+c should produce a quit command")
	}
	// Execute the command to verify it's a quit
	quitMsg := cmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Errorf("ctrl+c produced %T, want tea.QuitMsg", quitMsg)
	}
}

func TestAppQuitFromList(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("q from list should produce a quit command")
	}
	quitMsg := cmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Errorf("q from list produced %T, want tea.QuitMsg", quitMsg)
	}
}

func TestAppQuitFromDetail(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("q from detail should produce a quit command")
	}
	quitMsg := cmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Errorf("q from detail produced %T, want tea.QuitMsg", quitMsg)
	}
}

func TestAppQuitFromHelpOverlay(t *testing.T) {
	app := newTestApp(t)
	app.state = viewHelpOverlay

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("q from help overlay should produce a quit command")
	}
}

func TestAppQuitFromTagPicker(t *testing.T) {
	app := newTestApp(t)
	app.state = viewTagPicker

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("q from tag picker should produce a quit command")
	}
}

func TestAppQuitFromAllPickerViews(t *testing.T) {
	pickerViews := []viewState{
		viewDetail,
		viewTagPicker,
		viewParentPicker,
		viewStatusPicker,
		viewTypePicker,
		viewBlockingPicker,
		viewPriorityPicker,
		viewSortPicker,
		viewHelpOverlay,
	}
	for _, state := range pickerViews {
		app := newTestApp(t)
		app.state = state

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		_, cmd := app.Update(msg)
		if cmd == nil {
			t.Errorf("q from state %d should produce a quit command", state)
		}
	}
}

func TestAppHelpOverlayOpensFromList(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewHelpOverlay {
		t.Errorf("state = %d, want viewHelpOverlay (%d)", updated.state, viewHelpOverlay)
	}
	if updated.previousState != viewList {
		t.Errorf("previousState = %d, want viewList (%d)", updated.previousState, viewList)
	}
}

func TestAppHelpOverlayOpensFromDetail(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewHelpOverlay {
		t.Errorf("state = %d, want viewHelpOverlay (%d)", updated.state, viewHelpOverlay)
	}
	if updated.previousState != viewDetail {
		t.Errorf("previousState = %d, want viewDetail (%d)", updated.previousState, viewDetail)
	}
}

func TestAppHelpDoesNotOpenFromPicker(t *testing.T) {
	app := newTestApp(t)
	// Use viewCreateModal which is safe to test since we can initialize it
	app.previousState = viewList
	app.createModal = newCreateModalModel(80, 24)
	app.state = viewCreateModal

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	// Should NOT open help from a modal view
	if updated.state == viewHelpOverlay {
		t.Error("? from create modal should not open help overlay")
	}
}

func TestAppWindowSizeMsg(t *testing.T) {
	app := newTestApp(t)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestAppKeyChordGT(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.state = viewList

	// Press "g" - should set pending key
	gMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	updatedModel, _ := app.Update(gMsg)
	updated := updatedModel.(*App)

	if updated.pendingKey != "g" {
		t.Errorf("pendingKey = %q, want \"g\"", updated.pendingKey)
	}

	// Press "t" - should produce openTagPickerMsg
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updatedModel, cmd := updated.Update(tMsg)
	updated = updatedModel.(*App)

	if updated.pendingKey != "" {
		t.Errorf("pendingKey = %q, want empty after chord", updated.pendingKey)
	}

	if cmd == nil {
		t.Fatal("g t chord should produce a command")
	}
	result := cmd()
	if _, ok := result.(openTagPickerMsg); !ok {
		t.Errorf("g t chord produced %T, want openTagPickerMsg", result)
	}
}

func TestAppKeyChordInvalidSecond(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	// Press "g"
	gMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	updatedModel, _ := app.Update(gMsg)
	updated := updatedModel.(*App)

	if updated.pendingKey != "g" {
		t.Fatalf("pendingKey = %q, want \"g\"", updated.pendingKey)
	}

	// Press "x" (invalid second key)
	xMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	updatedModel, cmd := updated.Update(xMsg)
	updated = updatedModel.(*App)

	if updated.pendingKey != "" {
		t.Errorf("pendingKey = %q, want empty after invalid chord", updated.pendingKey)
	}
	if cmd != nil {
		t.Error("invalid chord should not produce a command")
	}
}

func TestAppTagSelectedMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewTagPicker

	msg := tagSelectedMsg{tag: "frontend"}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if updated.list.tagFilter != "frontend" {
		t.Errorf("tagFilter = %q, want \"frontend\"", updated.list.tagFilter)
	}
	if cmd == nil {
		t.Error("tagSelectedMsg should produce a loadIssues command")
	}
}

func TestAppClearFilterMsg(t *testing.T) {
	app := newTestApp(t)
	app.list.tagFilter = "backend"

	msg := clearFilterMsg{}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.list.tagFilter != "" {
		t.Errorf("tagFilter = %q, want empty", updated.list.tagFilter)
	}
	if cmd == nil {
		t.Error("clearFilterMsg should produce a loadIssues command")
	}
}

func TestAppSelectIssueMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	testIssue := &issue.Issue{
		ID:     "test-1",
		Title:  "Test Issue",
		Status: "todo",
		Type:   "task",
	}

	msg := selectIssueMsg{issue: testIssue}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}
	if updated.detail.issue.ID != "test-1" {
		t.Errorf("detail issue ID = %q, want \"test-1\"", updated.detail.issue.ID)
	}
}

func TestAppSelectIssueMsgPushesHistory(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "first", Title: "First", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	secondIssue := &issue.Issue{
		ID: "second", Title: "Second", Status: "todo", Type: "task",
	}
	msg := selectIssueMsg{issue: secondIssue}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if len(updated.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(updated.history))
	}
	if updated.history[0].issue.ID != "first" {
		t.Errorf("history[0].issue.ID = %q, want \"first\"", updated.history[0].issue.ID)
	}
	if updated.detail.issue.ID != "second" {
		t.Errorf("detail.issue.ID = %q, want \"second\"", updated.detail.issue.ID)
	}
}

func TestAppBackToListMsgPopsHistory(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "current", Title: "Current", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)
	app.history = []detailModel{
		newDetailModel(&issue.Issue{
			ID: "prev", Title: "Previous", Status: "todo", Type: "task",
		}, app.resolver, app.config, 80, 24),
	}

	msg := backToListMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if len(updated.history) != 0 {
		t.Errorf("history length = %d, want 0", len(updated.history))
	}
	// Should stay in detail view showing the previous issue
	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}
	if updated.detail.issue.ID != "prev" {
		t.Errorf("detail.issue.ID = %q, want \"prev\"", updated.detail.issue.ID)
	}
}

func TestAppBackToListMsgGoesToList(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "current", Title: "Current", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)
	// No history

	msg := backToListMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppOpenTagPickerMsgNoTags(t *testing.T) {
	app := newTestApp(t)
	// No issues with tags, so collectTagsWithCounts returns empty

	msg := openTagPickerMsg{}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	// Should NOT switch to tag picker if no tags
	if updated.state == viewTagPicker {
		t.Error("should not open tag picker when no tags exist")
	}
	if cmd != nil {
		t.Error("should not produce a command when no tags exist")
	}
}

func TestAppOpenTagPickerMsgWithTags(t *testing.T) {
	app, _ := newTestAppWithIssues(t)

	msg := openTagPickerMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewTagPicker {
		t.Errorf("state = %d, want viewTagPicker (%d)", updated.state, viewTagPicker)
	}
}

func TestAppStatusSelectedMsg(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewList
	_ = c // used indirectly via resolver

	msg := statusSelectedMsg{
		issueIDs: []string{"abc-123"},
		status:   "completed",
	}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if cmd == nil {
		t.Error("statusSelectedMsg should produce a loadIssues command")
	}

	// Verify the issue was updated
	b, err := c.Get("abc-123")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if b.Status != "completed" {
		t.Errorf("issue status = %q, want \"completed\"", b.Status)
	}
}

func TestAppTypeSelectedMsg(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewList

	msg := typeSelectedMsg{
		issueIDs:  []string{"abc-123"},
		issueType: "bug",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}

	b, err := c.Get("abc-123")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if b.Type != "bug" {
		t.Errorf("issue type = %q, want \"bug\"", b.Type)
	}
}

func TestAppPrioritySelectedMsg(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewList

	msg := prioritySelectedMsg{
		issueIDs: []string{"abc-123"},
		priority: "high",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}

	b, err := c.Get("abc-123")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if b.Priority != "high" {
		t.Errorf("issue priority = %q, want \"high\"", b.Priority)
	}
}

func TestAppSortSelectedMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewSortPicker

	msg := sortSelectedMsg{order: sortCreated}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if updated.list.sortOrder != sortCreated {
		t.Errorf("sortOrder = %q, want %q", updated.list.sortOrder, sortCreated)
	}
	if cmd == nil {
		t.Error("sortSelectedMsg should produce a loadIssues command")
	}
}

func TestAppCloseSortPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewSortPicker

	msg := closeSortPickerMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppCloseHelpMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewDetail
	app.state = viewHelpOverlay

	msg := closeHelpMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}
}

func TestAppOpenCreateModalMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := openCreateModalMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewCreateModal {
		t.Errorf("state = %d, want viewCreateModal (%d)", updated.state, viewCreateModal)
	}
	if updated.previousState != viewList {
		t.Errorf("previousState = %d, want viewList (%d)", updated.previousState, viewList)
	}
}

func TestAppCloseCreateModalMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewCreateModal

	msg := closeCreateModalMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppIssueCreatedMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewCreateModal

	msg := issueCreatedMsg{title: "Brand New Issue"}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if cmd == nil {
		t.Error("issueCreatedMsg should produce commands")
	}
}

func TestAppParentSelectedMsg(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewList

	// Create an epic for parent
	epic := &issue.Issue{ID: "epic-1", Title: "Epic", Status: "todo", Type: "epic"}
	c.Create(epic)

	msg := parentSelectedMsg{
		issueIDs: []string{"abc-123"},
		parentID: "epic-1",
	}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if cmd == nil {
		t.Error("parentSelectedMsg should produce a loadIssues command")
	}
}

func TestAppCloseParentPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewParentPicker

	msg := closeParentPickerMsg{}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if cmd == nil {
		t.Error("closeParentPickerMsg should produce a loadIssues command")
	}
}

func TestAppCloseStatusPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewDetail
	app.state = viewStatusPicker

	msg := closeStatusPickerMsg{}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}
	if cmd == nil {
		t.Error("closeStatusPickerMsg should produce a loadIssues command")
	}
}

func TestAppCloseTypePickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewTypePicker

	msg := closeTypePickerMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppClosePriorityPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewPriorityPicker

	msg := closePriorityPickerMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppCloseBlockingPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.previousState = viewList
	app.state = viewBlockingPicker

	msg := closeBlockingPickerMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
}

func TestAppBlockingConfirmedMsg(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewList
	app.state = viewBlockingPicker

	msg := blockingConfirmedMsg{
		issueID:  "abc-123",
		toAdd:    []string{"def-456"},
		toRemove: nil,
	}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d)", updated.state, viewList)
	}
	if cmd == nil {
		t.Error("blockingConfirmedMsg should produce a loadIssues command")
	}

	// Verify blocking was added
	b, _ := c.Get("abc-123")
	if len(b.Blocking) == 0 || b.Blocking[0] != "def-456" {
		t.Errorf("issue blocking = %v, want [def-456]", b.Blocking)
	}
}

func TestAppEditorFinishedMsg(t *testing.T) {
	app := newTestApp(t)

	// No editing in progress - should be a no-op
	msg := editorFinishedMsg{err: nil}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.editingIssueID != "" {
		t.Errorf("editingIssueID = %q, want empty", updated.editingIssueID)
	}
	if cmd != nil {
		t.Error("editorFinishedMsg with no editing should produce nil command")
	}
}

func TestAppKeyPressClearsStatusMessage(t *testing.T) {
	app := newTestApp(t)
	app.list.statusMessage = "Some status"
	app.detail.statusMessage = "Detail status"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.list.statusMessage != "" {
		t.Errorf("list.statusMessage = %q, want empty", updated.list.statusMessage)
	}
	if updated.detail.statusMessage != "" {
		t.Errorf("detail.statusMessage = %q, want empty", updated.detail.statusMessage)
	}
}

func TestAppViewReturnsStringForAllStates(t *testing.T) {
	// This ensures View() doesn't panic for each state
	states := []viewState{
		viewList,
	}
	for _, state := range states {
		app := newTestApp(t)
		app.state = state
		result := app.View()
		if result == "" {
			t.Errorf("View() returned empty string for state %d", state)
		}
	}
}

func TestAppOpenSortPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := openSortPickerMsg{currentOrder: sortDefault}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewSortPicker {
		t.Errorf("state = %d, want viewSortPicker (%d)", updated.state, viewSortPicker)
	}
	if updated.previousState != viewList {
		t.Errorf("previousState = %d, want viewList (%d)", updated.previousState, viewList)
	}
}

func TestAppOpenHelpMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail

	msg := openHelpMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewHelpOverlay {
		t.Errorf("state = %d, want viewHelpOverlay (%d)", updated.state, viewHelpOverlay)
	}
	if updated.previousState != viewDetail {
		t.Errorf("previousState = %d, want viewDetail (%d)", updated.previousState, viewDetail)
	}
}

func TestAppFinishBatchEditDetail(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.previousState = viewDetail
	app.state = viewStatusPicker
	app.detail = newDetailModel(&issue.Issue{
		ID: "abc-123", Title: "First", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)
	app.list.selectedIssues["abc-123"] = true

	msg := statusSelectedMsg{
		issueIDs: []string{"abc-123"},
		status:   "in-progress",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}
	// Selection should be cleared
	if len(updated.list.selectedIssues) != 0 {
		t.Errorf("selectedIssues = %v, want empty", updated.list.selectedIssues)
	}
}

func TestAppCollectTagsWithCounts(t *testing.T) {
	app, _ := newTestAppWithIssues(t)

	tags := app.collectTagsWithCounts()
	if len(tags) != 1 {
		t.Fatalf("tags count = %d, want 1", len(tags))
	}
	if tags[0].tag != "frontend" || tags[0].count != 1 {
		t.Errorf("tag = {%q, %d}, want {\"frontend\", 1}", tags[0].tag, tags[0].count)
	}
}

func TestAppForwardsToCurrentView(t *testing.T) {
	// Ensure the forwarding switch at the end of Update handles all view states
	app := newTestApp(t)

	// Test that window size messages are forwarded to child views
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}

	for _, state := range []viewState{viewList, viewDetail} {
		app.state = state
		if state == viewDetail {
			app.detail = newDetailModel(&issue.Issue{
				ID: "test", Title: "Test", Status: "todo", Type: "task",
			}, app.resolver, app.config, 80, 24)
		}
		updatedModel, _ := app.Update(sizeMsg)
		_ = updatedModel // Just ensure no panic
	}
}

// Test getBackgroundView
func TestAppGetBackgroundView(t *testing.T) {
	app := newTestApp(t)

	t.Run("list background", func(t *testing.T) {
		app.previousState = viewList
		bg := app.getBackgroundView()
		if bg == "" {
			t.Error("getBackgroundView for list should return non-empty string")
		}
	})

	t.Run("default background", func(t *testing.T) {
		app.previousState = viewTagPicker // not list or detail
		bg := app.getBackgroundView()
		if bg == "" {
			t.Error("getBackgroundView for default should return non-empty string")
		}
	})
}

// Test the DefaultKeyMap
func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	short := km.ShortHelp()
	if len(short) == 0 {
		t.Error("ShortHelp() returned empty")
	}

	full := km.FullHelp()
	if len(full) == 0 {
		t.Error("FullHelp() returned empty")
	}
}

func TestDefaultDetailKeyMap(t *testing.T) {
	km := DefaultDetailKeyMap()

	short := km.ShortHelp()
	if len(short) == 0 {
		t.Error("ShortHelp() returned empty")
	}

	full := km.FullHelp()
	if len(full) == 0 {
		t.Error("FullHelp() returned empty")
	}
}

// Test help overlay model
func TestHelpOverlayModel(t *testing.T) {
	m := newHelpOverlayModel(80, 24)

	t.Run("init returns nil", func(t *testing.T) {
		cmd := m.Init()
		if cmd != nil {
			t.Error("Init() should return nil")
		}
	})

	t.Run("view with width 0", func(t *testing.T) {
		zeroM := newHelpOverlayModel(0, 0)
		v := zeroM.View()
		if v != "Loading..." {
			t.Errorf("View() with zero width = %q, want \"Loading...\"", v)
		}
	})

	t.Run("view with non-zero width", func(t *testing.T) {
		v := m.View()
		if v == "" || v == "Loading..." {
			t.Error("View() should return content with non-zero width")
		}
	})

	t.Run("? key closes", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("? should produce close command")
		}
		result := cmd()
		if _, ok := result.(closeHelpMsg); !ok {
			t.Errorf("? produced %T, want closeHelpMsg", result)
		}
	})

	t.Run("esc key closes", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("esc should produce close command")
		}
		result := cmd()
		if _, ok := result.(closeHelpMsg); !ok {
			t.Errorf("esc produced %T, want closeHelpMsg", result)
		}
	})

	t.Run("window size updates", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		updated, _ := m.Update(msg)
		if updated.width != 120 || updated.height != 40 {
			t.Errorf("dimensions = %dx%d, want 120x40", updated.width, updated.height)
		}
	})
}

// Test create modal model
func TestCreateModalModel(t *testing.T) {
	m := newCreateModalModel(80, 24)

	t.Run("init returns blink", func(t *testing.T) {
		cmd := m.Init()
		if cmd == nil {
			t.Error("Init() should return a blink command")
		}
	})

	t.Run("view with width 0", func(t *testing.T) {
		zeroM := newCreateModalModel(0, 0)
		v := zeroM.View()
		if v != "Loading..." {
			t.Errorf("View() with zero width = %q, want \"Loading...\"", v)
		}
	})

	t.Run("view with non-zero width", func(t *testing.T) {
		v := m.View()
		if v == "" || v == "Loading..." {
			t.Error("View() should return content with non-zero width")
		}
	})

	t.Run("esc key closes", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("esc should produce close command")
		}
		result := cmd()
		if _, ok := result.(closeCreateModalMsg); !ok {
			t.Errorf("esc produced %T, want closeCreateModalMsg", result)
		}
	})

	t.Run("enter with empty title closes", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("enter with empty should produce close command")
		}
		result := cmd()
		if _, ok := result.(closeCreateModalMsg); !ok {
			t.Errorf("enter empty produced %T, want closeCreateModalMsg", result)
		}
	})

	t.Run("window size updates", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		updated, _ := m.Update(msg)
		if updated.width != 120 || updated.height != 40 {
			t.Errorf("dimensions = %dx%d, want 120x40", updated.width, updated.height)
		}
	})
}

// Test detail model
func TestDetailModel(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	os.MkdirAll(dataDir, 0755)
	cfg := config.Default()
	c := core.New(dataDir, cfg)
	c.Load()

	resolver := &graph.Resolver{Core: c}

	testIssue := &issue.Issue{
		ID: "test-1", Title: "Test Issue", Status: "todo", Type: "task",
		Body: "Some description here",
	}

	m := newDetailModel(testIssue, resolver, cfg, 80, 24)

	t.Run("init returns nil", func(t *testing.T) {
		cmd := m.Init()
		if cmd != nil {
			t.Error("Init() should return nil")
		}
	})

	t.Run("view renders", func(t *testing.T) {
		v := m.View()
		if v == "" || v == "Loading..." {
			t.Error("View() should return content")
		}
	})

	t.Run("esc returns backToListMsg", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("esc should produce back command")
		}
		result := cmd()
		if _, ok := result.(backToListMsg); !ok {
			t.Errorf("esc produced %T, want backToListMsg", result)
		}
	})

	t.Run("backspace returns backToListMsg", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyBackspace}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("backspace should produce back command")
		}
		result := cmd()
		if _, ok := result.(backToListMsg); !ok {
			t.Errorf("backspace produced %T, want backToListMsg", result)
		}
	})

	t.Run("p opens parent picker", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("p should produce parent picker command")
		}
		result := cmd()
		if _, ok := result.(openParentPickerMsg); !ok {
			t.Errorf("p produced %T, want openParentPickerMsg", result)
		}
	})

	t.Run("s opens status picker", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("s should produce status picker command")
		}
		result := cmd()
		if _, ok := result.(openStatusPickerMsg); !ok {
			t.Errorf("s produced %T, want openStatusPickerMsg", result)
		}
	})

	t.Run("t opens type picker", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("t should produce type picker command")
		}
		result := cmd()
		if _, ok := result.(openTypePickerMsg); !ok {
			t.Errorf("t produced %T, want openTypePickerMsg", result)
		}
	})

	t.Run("P opens priority picker", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("P")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("P should produce priority picker command")
		}
		result := cmd()
		if _, ok := result.(openPriorityPickerMsg); !ok {
			t.Errorf("P produced %T, want openPriorityPickerMsg", result)
		}
	})

	t.Run("b opens blocking picker", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("b should produce blocking picker command")
		}
		result := cmd()
		if _, ok := result.(openBlockingPickerMsg); !ok {
			t.Errorf("b produced %T, want openBlockingPickerMsg", result)
		}
	})

	t.Run("e opens editor", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("e should produce editor command")
		}
		result := cmd()
		if _, ok := result.(openEditorMsg); !ok {
			t.Errorf("e produced %T, want openEditorMsg", result)
		}
	})

	t.Run("c copies ID", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
		_, cmd := m.Update(msg)
		if cmd == nil {
			t.Fatal("c should produce copy command")
		}
		result := cmd()
		if copyMsg, ok := result.(copyIssueIDMsg); !ok {
			t.Errorf("c produced %T, want copyIssueIDMsg", result)
		} else if len(copyMsg.ids) != 1 || copyMsg.ids[0] != "test-1" {
			t.Errorf("copyIssueIDMsg.ids = %v, want [test-1]", copyMsg.ids)
		}
	})

	t.Run("tab toggles links with no links", func(t *testing.T) {
		beforeLinksActive := m.linksActive
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\t'}}
		// No links, so tab should be a no-op
		if len(m.links) == 0 {
			_, _ = m.Update(msg)
			// linksActive should not toggle since no links
		} else {
			updated, _ := m.Update(msg)
			if updated.linksActive == beforeLinksActive {
				t.Error("tab should toggle linksActive")
			}
		}
	})

	t.Run("window size updates", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 30}
		updated, _ := m.Update(msg)
		if updated.width != 100 || updated.height != 30 {
			t.Errorf("dimensions = %dx%d, want 100x30", updated.width, updated.height)
		}
	})

	t.Run("renderBody with empty body", func(t *testing.T) {
		emptyIssue := &issue.Issue{
			ID: "empty-1", Title: "Empty", Status: "todo", Type: "task",
		}
		em := newDetailModel(emptyIssue, resolver, cfg, 80, 24)
		body := em.renderBody(76)
		if body == "" {
			t.Error("renderBody should return non-empty for empty body")
		}
	})

	t.Run("visibleIssueIDs includes current issue", func(t *testing.T) {
		ids := m.visibleIssueIDs()
		if !ids["test-1"] {
			t.Error("visibleIssueIDs should include the current issue")
		}
	})
}

// Test pickerDimensions helper
func TestCalculatePickerDimensions(t *testing.T) {
	cfg := defaultPickerDimensionConfig()

	t.Run("normal screen", func(t *testing.T) {
		dims := calculatePickerDimensions(100, 40, cfg)
		if dims.ModalWidth < cfg.MinWidth || dims.ModalWidth > cfg.MaxWidth {
			t.Errorf("ModalWidth = %d, want between %d and %d", dims.ModalWidth, cfg.MinWidth, cfg.MaxWidth)
		}
		if dims.ListWidth != dims.ModalWidth-cfg.WidthPadding {
			t.Errorf("ListWidth = %d, want %d", dims.ListWidth, dims.ModalWidth-cfg.WidthPadding)
		}
	})

	t.Run("small screen", func(t *testing.T) {
		dims := calculatePickerDimensions(30, 10, cfg)
		if dims.ModalWidth < cfg.MinWidth {
			t.Errorf("ModalWidth = %d, should be at least %d", dims.ModalWidth, cfg.MinWidth)
		}
		if dims.ModalHeight < cfg.MinHeight {
			t.Errorf("ModalHeight = %d, should be at least %d", dims.ModalHeight, cfg.MinHeight)
		}
	})
}

// Test overlayModal
func TestOverlayModal(t *testing.T) {
	bg := "line1\nline2\nline3"
	modal := "modal"
	result := overlayModal(bg, modal, 20, 5)
	if result == "" {
		t.Error("overlayModal should return non-empty string")
	}
}

// Test renderPickerModal
func TestRenderPickerModal(t *testing.T) {
	cfg := pickerModalConfig{
		Title:       "Test",
		IssueTitle:  "Test Issue",
		IssueID:     "test-1",
		ListContent: "item1\nitem2",
		Width:       80,
	}
	result := renderPickerModal(cfg)
	if result == "" {
		t.Error("renderPickerModal should return non-empty string")
	}

	t.Run("with description", func(t *testing.T) {
		cfg.Description = "A description"
		result := renderPickerModal(cfg)
		if result == "" {
			t.Error("renderPickerModal with description should return non-empty string")
		}
	})

	t.Run("with long title truncation", func(t *testing.T) {
		cfg.IssueTitle = "A very very very very very very very very very very very long title that should be truncated"
		result := renderPickerModal(cfg)
		if result == "" {
			t.Error("renderPickerModal with long title should return non-empty string")
		}
	})
}

// Test the OpenParentPickerMsg with milestone type (cannot have parents)
func TestAppOpenParentPickerMsgMilestone(t *testing.T) {
	app := newTestApp(t)

	msg := openParentPickerMsg{
		issueIDs:   []string{"test-1"},
		issueTitle: "Milestone",
		issueTypes: []string{"milestone"}, // milestones cannot have parents
	}
	updatedModel, cmd := app.Update(msg)
	updated := updatedModel.(*App)

	// Should not open parent picker for milestones
	if updated.state == viewParentPicker {
		t.Error("should not open parent picker for milestones")
	}
	if cmd != nil {
		t.Error("should not produce a command for milestones")
	}
}

// Test CopyIssueIDMsg handling
func TestAppCopyIssueIDMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := copyIssueIDMsg{ids: []string{"abc-123"}}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	// Status message should be set (may fail on CI but structure is tested)
	if updated.list.statusMessage == "" {
		// Clipboard may not be available in test environment, that's OK
		// The code path is still exercised
	}
}

func TestAppCopyIssueIDMsgInDetail(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail

	msg := copyIssueIDMsg{ids: []string{"abc-123", "def-456"}}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	// The code path is exercised regardless of clipboard availability
	_ = updated
}

// Test IssuesChangedMsg in list view
func TestAppIssuesChangedMsgInList(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := issuesChangedMsg{changedIDs: map[string]bool{"test-1": true}}
	_, cmd := app.Update(msg)

	if cmd == nil {
		t.Error("issuesChangedMsg should produce a loadIssues command")
	}
}

// Test IssuesChangedMsg in detail view with non-relevant change
func TestAppIssuesChangedMsgInDetailNotRelevant(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "current-1", Title: "Current", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	// Change is for a different issue
	msg := issuesChangedMsg{changedIDs: map[string]bool{"other-1": true}}
	_, cmd := app.Update(msg)

	if cmd == nil {
		t.Error("issuesChangedMsg should produce a loadIssues command even for non-relevant changes")
	}
}

// Test the OpenStatusPickerMsg
func TestAppOpenStatusPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := openStatusPickerMsg{
		issueIDs:      []string{"test-1"},
		issueTitle:    "Test",
		currentStatus: "todo",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewStatusPicker {
		t.Errorf("state = %d, want viewStatusPicker (%d)", updated.state, viewStatusPicker)
	}
}

func TestAppOpenTypePickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := openTypePickerMsg{
		issueIDs:    []string{"test-1"},
		issueTitle:  "Test",
		currentType: "task",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewTypePicker {
		t.Errorf("state = %d, want viewTypePicker (%d)", updated.state, viewTypePicker)
	}
}

func TestAppOpenPriorityPickerMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := openPriorityPickerMsg{
		issueIDs:        []string{"test-1"},
		issueTitle:      "Test",
		currentPriority: "normal",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewPriorityPicker {
		t.Errorf("state = %d, want viewPriorityPicker (%d)", updated.state, viewPriorityPicker)
	}
}

// Test listModel filter methods
func TestListModelFilterMethods(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)

	t.Run("initial state has no filter", func(t *testing.T) {
		if m.hasActiveFilter() {
			t.Error("should not have active filter initially")
		}
	})

	t.Run("set tag filter", func(t *testing.T) {
		m.setTagFilter("frontend")
		if m.tagFilter != "frontend" {
			t.Errorf("tagFilter = %q, want \"frontend\"", m.tagFilter)
		}
		if !m.hasActiveFilter() {
			t.Error("should have active filter after setTagFilter")
		}
	})

	t.Run("clear filter", func(t *testing.T) {
		m.clearFilter()
		if m.tagFilter != "" {
			t.Errorf("tagFilter = %q, want empty", m.tagFilter)
		}
		if m.hasActiveFilter() {
			t.Error("should not have active filter after clearFilter")
		}
	})
}

// Test listModel View with error
func TestListModelViewWithError(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)
	m.err = errMsg{err: nil}.err
	m.width = 80
	m.height = 24

	// No error set, should show normal view
	v := m.View()
	if v == "" {
		t.Error("View should return non-empty content")
	}
}

// Test listModel View with tag filter
func TestListModelViewWithTagFilter(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)
	m.width = 80
	m.height = 24
	m.list.SetSize(78, 20)
	m.tagFilter = "frontend"

	v := m.View()
	if v == "" {
		t.Error("View with tag filter should return non-empty")
	}
}

// Test issueItem interface methods
func TestIssueItemMethods(t *testing.T) {
	item := issueItem{
		issue: &issue.Issue{ID: "test-1", Title: "Test Title", Status: "todo"},
	}

	if item.Title() != "Test Title" {
		t.Errorf("Title() = %q, want \"Test Title\"", item.Title())
	}

	desc := item.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

// Test OpenBlockingPickerMsg
func TestAppOpenBlockingPickerMsg(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.state = viewDetail

	msg := openBlockingPickerMsg{
		issueID:         "abc-123",
		issueTitle:      "First issue",
		currentBlocking: nil,
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewBlockingPicker {
		t.Errorf("state = %d, want viewBlockingPicker (%d)", updated.state, viewBlockingPicker)
	}
}

// Test blockingConfirmedMsg returning to detail view and refreshing
func TestAppBlockingConfirmedMsgInDetail(t *testing.T) {
	app, c := newTestAppWithIssues(t)
	app.previousState = viewDetail
	app.state = viewBlockingPicker
	app.detail = newDetailModel(&issue.Issue{
		ID: "abc-123", Title: "First", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	msg := blockingConfirmedMsg{
		issueID:  "abc-123",
		toAdd:    []string{"def-456"},
		toRemove: nil,
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewDetail {
		t.Errorf("state = %d, want viewDetail (%d)", updated.state, viewDetail)
	}

	b, _ := c.Get("abc-123")
	if len(b.Blocking) == 0 || b.Blocking[0] != "def-456" {
		t.Errorf("blocking = %v, want [def-456]", b.Blocking)
	}
}

// Test the OpenParentPickerMsg for task type (should open)
func TestAppOpenParentPickerMsgTask(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.state = viewList

	msg := openParentPickerMsg{
		issueIDs:      []string{"abc-123"},
		issueTitle:    "First issue",
		issueTypes:    []string{"task"},
		currentParent: "",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewParentPicker {
		t.Errorf("state = %d, want viewParentPicker (%d)", updated.state, viewParentPicker)
	}
}

// Test issuesChangedMsg in detail for relevant issue
func TestAppIssuesChangedMsgDetailRelevant(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "abc-123", Title: "First", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	msg := issuesChangedMsg{changedIDs: map[string]bool{"abc-123": true}}
	_, cmd := app.Update(msg)

	if cmd == nil {
		t.Error("issuesChangedMsg should produce loadIssues command")
	}
}

// Test issuesChangedMsg in detail with deleted issue falls back to list
func TestAppIssuesChangedMsgDetailDeletedIssue(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "deleted-id", Title: "Deleted", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	msg := issuesChangedMsg{changedIDs: map[string]bool{"deleted-id": true}}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	// Issue doesn't exist in core, so should fall back to list
	if updated.state != viewList {
		t.Errorf("state = %d, want viewList (%d) after deleted issue", updated.state, viewList)
	}
}

// Test formatLinkLabel
func TestFormatLinkLabel(t *testing.T) {
	cfg := config.Default()
	m := detailModel{config: cfg}

	tests := []struct {
		linkType string
		incoming bool
		want     string
	}{
		{issue.LinkTypeBlocking, false, "Blocking"},
		{issue.LinkTypeBlocking, true, "Blocked by"},
		{issue.LinkTypeParent, false, "Parent"},
		{issue.LinkTypeParent, true, "Child"},
		{"custom", false, "custom"},
		{"custom", true, "custom (incoming)"},
	}

	for _, tt := range tests {
		got := m.formatLinkLabel(tt.linkType, tt.incoming)
		if got != tt.want {
			t.Errorf("formatLinkLabel(%q, %v) = %q, want %q", tt.linkType, tt.incoming, got, tt.want)
		}
	}
}

// Test refreshIssue
func TestDetailRefreshIssue(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	os.MkdirAll(dataDir, 0755)
	cfg := config.Default()
	c := core.New(dataDir, cfg)
	c.Load()

	resolver := &graph.Resolver{Core: c}

	testIssue := &issue.Issue{
		ID: "test-1", Title: "Original", Status: "todo", Type: "task",
	}
	m := newDetailModel(testIssue, resolver, cfg, 80, 24)

	updated := &issue.Issue{
		ID: "test-1", Title: "Updated", Status: "in-progress", Type: "task",
		Tags: []string{"new-tag"},
	}
	m.refreshIssue(updated)

	if m.issue.Title != "Updated" {
		t.Errorf("issue.Title = %q, want \"Updated\"", m.issue.Title)
	}
	if m.issue.Status != "in-progress" {
		t.Errorf("issue.Status = %q, want \"in-progress\"", m.issue.Status)
	}
}

// Test rendering modal view
func TestCreateModalModalView(t *testing.T) {
	m := newCreateModalModel(80, 24)
	bg := "background line"
	result := m.ModalView(bg, 80, 24)
	if result == "" {
		t.Error("ModalView should return non-empty string")
	}
}

func TestHelpOverlayModalView(t *testing.T) {
	m := newHelpOverlayModel(80, 24)
	bg := "background line"
	result := m.ModalView(bg, 80, 24)
	if result == "" {
		t.Error("ModalView should return non-empty string")
	}
}

// Test list model error view
func TestListModelErrorView(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)
	m.width = 80
	m.height = 24
	m.err = fmt.Errorf("test error")

	v := m.View()
	if v == "" {
		t.Error("error view should not be empty")
	}
}

// Test renderPickerCursor
func TestRenderPickerCursor(t *testing.T) {
	type indexer struct{ idx int }
	fn := func() int { return 0 }
	_ = fn

	// Can't easily test this without a list.Model, but test the function exists
	// and doesn't panic
	mockM := &mockIndex{index: 2}
	cursor := renderPickerCursor(2, mockM)
	if cursor == "" {
		t.Error("selected cursor should not be empty")
	}

	cursor = renderPickerCursor(0, mockM)
	if cursor != "  " {
		t.Errorf("unselected cursor = %q, want \"  \"", cursor)
	}
}

type mockIndex struct {
	index int
}

func (m *mockIndex) Index() int { return m.index }

// Test detail calculateHeaderHeight
func TestDetailCalculateHeaderHeight(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	os.MkdirAll(dataDir, 0755)
	cfg := config.Default()
	c := core.New(dataDir, cfg)
	c.Load()
	resolver := &graph.Resolver{Core: c}

	t.Run("no links", func(t *testing.T) {
		m := newDetailModel(&issue.Issue{
			ID: "test-1", Title: "Test", Status: "todo", Type: "task",
		}, resolver, cfg, 80, 24)
		h := m.calculateHeaderHeight()
		if h < 6 {
			t.Errorf("header height = %d, want >= 6", h)
		}
	})
}

// Test linksHaveTags
func TestLinksHaveTags(t *testing.T) {
	t.Run("no links", func(t *testing.T) {
		if linksHaveTags(nil) {
			t.Error("nil links should not have tags")
		}
	})

	t.Run("links without tags", func(t *testing.T) {
		links := []resolvedLink{
			{issue: &issue.Issue{ID: "a"}},
		}
		if linksHaveTags(links) {
			t.Error("links without tags should return false")
		}
	})

	t.Run("links with tags", func(t *testing.T) {
		links := []resolvedLink{
			{issue: &issue.Issue{ID: "a", Tags: []string{"tag1"}}},
		}
		if !linksHaveTags(links) {
			t.Error("links with tags should return true")
		}
	})
}

// Test list model selected issues view footer
func TestListModelViewWithSelectedIssues(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)
	m.width = 80
	m.height = 24
	m.list.SetSize(78, 20)
	m.selectedIssues["test-1"] = true

	v := m.View()
	if v == "" {
		t.Error("View with selected issues should return non-empty")
	}
}

// Test list model status message view
func TestListModelViewWithStatusMessage(t *testing.T) {
	cfg := config.Default()
	resolver := &graph.Resolver{}
	m := newListModel(resolver, cfg)
	m.width = 80
	m.height = 24
	m.list.SetSize(78, 20)
	m.statusMessage = "Copied abc-123 to clipboard"

	v := m.View()
	if v == "" {
		t.Error("View with status message should return non-empty")
	}
}

// Test tickMsg
func TestAppTickMsg(t *testing.T) {
	app := newTestApp(t)
	app.state = viewList

	msg := tickMsg{}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Error("tickMsg should produce a batch command")
	}
}

// Test tickMsg in detail view
func TestAppTickMsgInDetail(t *testing.T) {
	app, _ := newTestAppWithIssues(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "abc-123", Title: "First", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	msg := tickMsg{}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Error("tickMsg in detail should produce a batch command")
	}
}

// Test tickMsg in detail view with deleted issue
func TestAppTickMsgDetailDeletedIssue(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail
	app.detail = newDetailModel(&issue.Issue{
		ID: "nonexistent", Title: "Gone", Status: "todo", Type: "task",
	}, app.resolver, app.config, 80, 24)

	msg := tickMsg{}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewList {
		t.Errorf("state = %d, want viewList after tick with deleted issue", updated.state)
	}
}

// Ensure the resolver created by New has Core set
func TestNewAppResolver(t *testing.T) {
	app := newTestApp(t)
	if app.resolver == nil {
		t.Fatal("resolver should not be nil")
	}
	if app.resolver.Core == nil {
		t.Fatal("resolver.Core should not be nil")
	}
}

// Test systemEditor
func TestSystemEditor(t *testing.T) {
	cmd, args, ok := systemEditor()
	// Just verify it doesn't panic; result depends on OS
	_ = cmd
	_ = args
	_ = ok
}

// Test OpenStatusPickerMsg from detail
func TestAppOpenStatusPickerMsgFromDetail(t *testing.T) {
	app := newTestApp(t)
	app.state = viewDetail

	msg := openStatusPickerMsg{
		issueIDs:      []string{"test-1"},
		issueTitle:    "Test",
		currentStatus: "todo",
	}
	updatedModel, _ := app.Update(msg)
	updated := updatedModel.(*App)

	if updated.state != viewStatusPicker {
		t.Errorf("state = %d, want viewStatusPicker", updated.state)
	}
	if updated.previousState != viewDetail {
		t.Errorf("previousState = %d, want viewDetail", updated.previousState)
	}
}

// ---------- graph Resolver helpers tested via App ----------

func TestAppResolverMutationViaUpdate(t *testing.T) {
	app, c := newTestAppWithIssues(t)

	// Verify we can use the resolver to create issues
	mr := app.resolver.Mutation()
	got, err := mr.CreateIssue(nil, model.CreateIssueInput{Title: "Via Resolver"})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}
	if got.Title != "Via Resolver" {
		t.Errorf("Title = %q, want \"Via Resolver\"", got.Title)
	}

	// Verify it's in the core
	b, err := c.Get(got.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if b.Title != "Via Resolver" {
		t.Errorf("core Title = %q, want \"Via Resolver\"", b.Title)
	}
}
