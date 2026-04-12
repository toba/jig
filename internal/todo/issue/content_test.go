package issue

import (
	"strings"
	"testing"
)

func TestReplaceOnce(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		old     string
		new     string
		want    string
		wantErr string
	}{
		{
			name: "simple replacement",
			text: "hello world",
			old:  "world",
			new:  "there",
			want: "hello there",
		},
		{
			name: "replace checkbox unchecked to checked",
			text: "## Tasks\n- [ ] Task 1\n- [ ] Task 2",
			old:  "- [ ] Task 1",
			new:  "- [x] Task 1",
			want: "## Tasks\n- [x] Task 1\n- [ ] Task 2",
		},
		{
			name: "delete text with empty new",
			text: "hello world",
			old:  " world",
			new:  "",
			want: "hello",
		},
		{
			name: "replace at start",
			text: "hello world",
			old:  "hello",
			new:  "hi",
			want: "hi world",
		},
		{
			name: "replace at end",
			text: "hello world",
			old:  "world",
			new:  "universe",
			want: "hello universe",
		},
		{
			name: "replace entire string",
			text: "hello",
			old:  "hello",
			new:  "goodbye",
			want: "goodbye",
		},
		{
			name: "replace with longer text",
			text: "a",
			old:  "a",
			new:  "abc",
			want: "abc",
		},
		{
			name: "replace multiline",
			text: "line1\nline2\nline3",
			old:  "line2",
			new:  "replaced",
			want: "line1\nreplaced\nline3",
		},
		{
			name:    "empty old string",
			text:    "hello",
			old:     "",
			new:     "world",
			wantErr: "old text cannot be empty",
		},
		{
			name:    "text not found",
			text:    "hello world",
			old:     "foo",
			new:     "bar",
			wantErr: "text not found in body",
		},
		{
			name:    "text found multiple times",
			text:    "hello hello",
			old:     "hello",
			new:     "hi",
			wantErr: "text found 2 times in body (must be unique)",
		},
		{
			name:    "text found three times",
			text:    "aaa",
			old:     "a",
			new:     "b",
			wantErr: "text found 3 times in body (must be unique)",
		},
		{
			name:    "empty text with non-empty old",
			text:    "",
			old:     "hello",
			new:     "world",
			wantErr: "text not found in body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceOnce(tt.text, tt.old, tt.new)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("ReplaceOnce() error = nil, wantErr %q", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("ReplaceOnce() error = %q, wantErr %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ReplaceOnce() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceOnce() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckItem(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		substr  string
		want    string
		wantErr string
	}{
		{
			name:   "check by exact label",
			text:   "- [ ] Task 1\n- [ ] Task 2",
			substr: "Task 1",
			want:   "- [x] Task 1\n- [ ] Task 2",
		},
		{
			name:   "check by partial substring",
			text:   "- [ ] Investigate fuzzy matching\n- [ ] Add tests",
			substr: "fuzzy",
			want:   "- [x] Investigate fuzzy matching\n- [ ] Add tests",
		},
		{
			name:   "check item with backticks",
			text:   "- [ ] Add `debug_detach` call\n- [ ] Fix other thing",
			substr: "debug_detach",
			want:   "- [x] Add `debug_detach` call\n- [ ] Fix other thing",
		},
		{
			name:   "check last item",
			text:   "- [x] Done\n- [ ] Still todo",
			substr: "Still todo",
			want:   "- [x] Done\n- [x] Still todo",
		},
		{
			name:   "case insensitive match",
			text:   "- [ ] Fix The Bug",
			substr: "fix the bug",
			want:   "- [x] Fix The Bug",
		},
		{
			name:    "no matching checkbox",
			text:    "- [ ] Task 1\n- [ ] Task 2",
			substr:  "nonexistent",
			wantErr: "no unchecked item matching",
		},
		{
			name:    "multiple matches",
			text:    "- [ ] Fix bug in parser\n- [ ] Fix bug in lexer",
			substr:  "Fix bug",
			wantErr: "2 unchecked items match",
		},
		{
			name:    "already checked",
			text:    "- [x] Task 1\n- [ ] Task 2",
			substr:  "Task 1",
			wantErr: "no unchecked item matching",
		},
		{
			name:    "empty substring",
			text:    "- [ ] Task 1",
			substr:  "",
			wantErr: "search text cannot be empty",
		},
		{
			name:   "with surrounding non-checkbox content",
			text:   "## Tasks\n\n- [ ] Do the thing\n\nSome notes",
			substr: "Do the thing",
			want:   "## Tasks\n\n- [x] Do the thing\n\nSome notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckItem(tt.text, tt.substr)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("CheckItem() error = nil, wantErr containing %q", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("CheckItem() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("CheckItem() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("CheckItem() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUncheckItem(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		substr  string
		want    string
		wantErr string
	}{
		{
			name:   "uncheck by exact label",
			text:   "- [x] Task 1\n- [x] Task 2",
			substr: "Task 1",
			want:   "- [ ] Task 1\n- [x] Task 2",
		},
		{
			name:   "uncheck by partial substring",
			text:   "- [x] Investigate fuzzy matching\n- [x] Add tests",
			substr: "fuzzy",
			want:   "- [ ] Investigate fuzzy matching\n- [x] Add tests",
		},
		{
			name:    "not checked",
			text:    "- [ ] Task 1\n- [x] Task 2",
			substr:  "Task 1",
			wantErr: "no checked item matching",
		},
		{
			name:    "no match",
			text:    "- [x] Task 1",
			substr:  "nonexistent",
			wantErr: "no checked item matching",
		},
		{
			name:    "multiple matches",
			text:    "- [x] Fix bug A\n- [x] Fix bug B",
			substr:  "Fix bug",
			wantErr: "2 checked items match",
		},
		{
			name:    "empty substring",
			text:    "- [x] Task 1",
			substr:  "",
			wantErr: "search text cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UncheckItem(tt.text, tt.substr)
			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("UncheckItem() error = nil, wantErr containing %q", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("UncheckItem() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("UncheckItem() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("UncheckItem() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasIncompleteChecklist(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"empty body", "", false},
		{"no checkboxes", "Some plain text\nMore text", false},
		{"all checked", "- [x] Done\n- [x] Also done", false},
		{"one unchecked", "- [ ] Todo", true},
		{"mixed checked and unchecked", "- [x] Done\n- [ ] Todo", true},
		{"indented unchecked", "  - [ ] Nested todo", true},
		{"tab indented unchecked", "\t- [ ] Nested todo", true},
		{"unchecked with surrounding content", "## Tasks\n\n- [ ] Do thing\n\nSome notes", true},
		{"all checked with surrounding content", "## Tasks\n\n- [x] Done\n\nNotes", false},
		{"no trailing space not a checkbox", "- []not a checkbox", false},
		{"only bracket no space", "- [ ]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasIncompleteChecklist(tt.text)
			if got != tt.want {
				t.Errorf("HasIncompleteChecklist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendWithSeparator(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		addition string
		want     string
	}{
		{
			name:     "append to non-empty text",
			text:     "hello",
			addition: "world",
			want:     "hello\n\nworld",
		},
		{
			name:     "append to empty text",
			text:     "",
			addition: "world",
			want:     "world",
		},
		{
			name:     "append empty to non-empty text (no-op)",
			text:     "hello",
			addition: "",
			want:     "hello",
		},
		{
			name:     "append empty to empty text (no-op)",
			text:     "",
			addition: "",
			want:     "",
		},
		{
			name:     "text with trailing newline",
			text:     "hello\n",
			addition: "world",
			want:     "hello\n\nworld",
		},
		{
			name:     "text with multiple trailing newlines",
			text:     "hello\n\n\n",
			addition: "world",
			want:     "hello\n\nworld",
		},
		{
			name:     "multiline text",
			text:     "line1\nline2",
			addition: "line3",
			want:     "line1\nline2\n\nline3",
		},
		{
			name:     "multiline addition",
			text:     "header",
			addition: "line1\nline2",
			want:     "header\n\nline1\nline2",
		},
		{
			name:     "typical usage - adding notes section",
			text:     "## Tasks\n- [ ] Task 1",
			addition: "## Notes\n\nSome notes here",
			want:     "## Tasks\n- [ ] Task 1\n\n## Notes\n\nSome notes here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppendWithSeparator(tt.text, tt.addition)
			if got != tt.want {
				t.Errorf("AppendWithSeparator() = %q, want %q", got, tt.want)
			}
		})
	}
}
