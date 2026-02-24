---
# krz-eaz
title: Fix all golangci-lint issues
status: in-progress
type: task
priority: normal
created_at: 2026-02-24T17:35:30Z
updated_at: 2026-02-24T17:36:02Z
---

Work through all 122 remaining golangci-lint issues that require manual intervention. Run `./scripts/lint.sh` to verify progress.

## errcheck (50)

### Production code
- [ ] `internal/brew/init.go:212` — `defer os.RemoveAll(tmp)`
- [ ] `internal/brew/sha.go:68` — `defer os.RemoveAll(tmp)`
- [ ] `internal/cite/add.go:97` — `defer os.RemoveAll(tmpDir)`
- [ ] `internal/display/display.go:55` — `fmt.Fprintln(w, sepStyle.Render("---"))`
- [ ] `internal/display/display.go:56` — `fmt.Fprintln(w)`
- [ ] `internal/display/display.go:62` — `fmt.Fprintln(w, header)`
- [ ] `internal/display/display.go:82` — `fmt.Fprintf(w, ...)`
- [ ] `internal/nope/debug.go:44` — `d.f.Write(data)`
- [ ] `internal/nope/debug.go:45` — `d.f.WriteString("\n")`
- [ ] `internal/nope/debug.go:53` — `d.f.Close()`
- [ ] `internal/testutil/chdir.go:17` — `os.Chdir(orig)` in cleanup
- [ ] `internal/todo/core/core.go:92` — `fmt.Fprintf(c.warnWriter, ...)`
- [ ] `internal/todo/core/core.go:145` — `c.searchIndex.Close()`
- [ ] `internal/todo/core/core.go:162` — `defer f.Close()`
- [ ] `internal/todo/core/watcher.go:128` — `watcher.Close()`
- [ ] `internal/todo/core/watcher.go:184` — `defer watcher.Close()`
- [ ] `internal/todo/integration/syncutil/images.go:92` — `defer f.Close()`

### Test code
- [ ] `internal/cite/doctor_test.go:82` — `os.WriteFile("LICENSE", ...)`
- [ ] `internal/cite/doctor_test.go:83` — `os.WriteFile("NOTICE", ...)`
- [ ] `internal/cite/doctor_test.go:119` — `os.WriteFile("LICENSE", ...)`
- [ ] `internal/nope/guard_test.go:72` — `w.Close()`
- [ ] `internal/todo/core/core_test.go:399` — `os.Mkdir(..., "subdir")`
- [ ] `internal/todo/core/core_test.go:649` — `defer core.Unwatch()`
- [ ] `internal/todo/core/core_test.go:697` — `defer core.Unwatch()`
- [ ] `internal/todo/core/core_test.go:740` — `defer core.Unwatch()`
- [ ] `internal/todo/core/core_test.go:884` — `os.Remove(...)`
- [ ] `internal/todo/core/search_test.go:12` — `defer core.Close()`
- [ ] `internal/todo/core/search_test.go:40` — `defer core.Close()`
- [ ] `internal/todo/core/search_test.go:65` — `defer core.Close()`
- [ ] `internal/todo/core/watcher_test.go:31` — `defer watcher.Close()`
- [ ] `internal/todo/integration/github/images_test.go:19` — `json.NewEncoder(w).Encode(...)`
- [ ] `internal/todo/integration/github/images_test.go:40` — `json.NewEncoder(w).Encode(...)`
- [ ] `internal/todo/integration/github/images_test.go:64` — `json.NewEncoder(w).Encode(...)`
- [ ] `internal/todo/integration/github/sync_test.go:979` — `fmt.Sscanf(...)`
- [ ] `internal/todo/integration/github/sync_test.go:1052` — `fmt.Sscanf(...)`
- [ ] `internal/todo/integration/github/sync_test.go:1055` — `fmt.Sscanf(...)`
- [ ] `internal/todo/issue/issue_test.go:2209` — `json.Unmarshal(data1, ...)`
- [ ] `internal/todo/issue/issue_test.go:2210` — `json.Unmarshal(data2, ...)`
- [ ] `internal/todo/output/output_test.go:26` — `w.Close()`
- [ ] `internal/todo/output/output_test.go:31` — `r.Close()`
- [ ] `internal/todo/search/index_test.go:17` — `idx.Close()`
- [ ] `internal/todo/search/index_test.go:27` — `defer idx.Close()`
- [ ] `internal/update/commit_test.go:64` — `os.Chdir(orig)` in cleanup
- [ ] `internal/update/commit_test.go:65` — `os.Chdir(tmp)`
- [ ] `internal/zed/doctor_test.go:197` — `os.Setenv("PATH", ...)`
- [ ] `internal/zed/doctor_test.go:198` — `defer os.Setenv("PATH", ...)`
- [ ] `internal/zed/doctor_test.go:207` — `os.MkdirAll(".github/workflows", ...)`
- [ ] `internal/zed/doctor_test.go:248` — `os.Setenv("PATH", ...)`
- [ ] `internal/zed/doctor_test.go:290` — `os.MkdirAll(".github/workflows", ...)`
- [ ] `internal/zed/doctor_test.go:332` — `os.MkdirAll(".github/workflows", ...)`

## gosec (21)
- [ ] `internal/commit/commit.go:82` — G204: subprocess with tainted input (`git log tag+..HEAD`)
- [ ] `internal/commit/commit_test.go:84` — G204: subprocess with variable
- [ ] `internal/commit/commit_test.go:256` — G204: subprocess with variable
- [ ] `internal/companion/gh.go:29` — G204: subprocess with variable (`gh` args)
- [ ] `internal/companion/gh.go:47` — G204: subprocess with variable (`gh release list`)
- [ ] `internal/config/companions.go:49` — G306: WriteFile 0o644 perms
- [ ] `internal/config/config.go:43` — G304: file inclusion via variable
- [ ] `internal/config/config.go:101` — G306: WriteFile 0o644 perms
- [ ] `internal/config/config.go:208` — G306: WriteFile 0o644 perms
- [ ] `internal/config/config_test.go:110` — G304: file inclusion via variable
- [ ] `internal/nope/config.go:24` — G101: `BuiltinCredentialRead` const name
- [ ] `internal/testutil/readfile.go:10` — G304: file inclusion via variable
- [ ] `internal/todo/config/config_test.go:522` — G301: MkdirAll 0755 perms
- [ ] `internal/todo/core/core.go:422` — G104: unhandled `h.Write(content)`
- [ ] `internal/todo/integration/syncutil/retry.go:70` — G404: weak random (`rand.Int64N`)
- [ ] `internal/todo/integration/syncutil/retry.go:90` — G704: SSRF via taint analysis
- [ ] `internal/todo/issue/issue.go:346` — G104: unhandled `h.Write(content)`
- [ ] `internal/update/commit.go:101` — G104: unhandled `os.Remove(dir)`
- [ ] `internal/update/commit_test.go:69` — G301: MkdirAll 0o755 perms
- [ ] `internal/update/commit_test.go:80` — G301: MkdirAll 0o755 perms
- [ ] `pkg/client/client_test.go:15` — G204: subprocess with variable

## gocritic (18)
- [ ] `cmd/init.go:45` — emptyStringTest: `len(content) > 0` → `content != ""`
- [ ] `cmd/todo_content.go:60` — paramTypeCombine: `code string, format string` → `code, format string`
- [ ] `internal/cite/add.go:259` — paramTypeCombine: `slice []string, files []string`
- [ ] `internal/nope/guard.go:71` — paramTypeCombine: `toolName string, input string`
- [ ] `internal/nope/init.go:132` — emptyStringTest: `len(content) > 0`
- [ ] `internal/todo/core/links.go:142` — appendAssign: append result not assigned to same slice
- [ ] `internal/todo/core/links.go:241` — appendAssign: append result not assigned to same slice
- [ ] `internal/todo/core/watcher.go:113` — deprecatedComment: needs dedicated paragraph
- [ ] `internal/todo/graph/schema.resolvers.go:359` — paramTypeCombine: `id string, name string`
- [ ] `internal/todo/graph/schema.resolvers.go:378` — paramTypeCombine: `id string, name string`
- [ ] `internal/todo/integration/clickup/config.go:142` — singleCaseSwitch: rewrite to if
- [ ] `internal/todo/integration/github/sync.go:475` — appendCombine: combine 2 appends
- [ ] `internal/todo/output/output.go:94` — paramTypeCombine: `code string, message string`
- [ ] `internal/todo/tui/tui.go:436` — appendAssign: append result not assigned to same slice
- [ ] `internal/todo/tui/tui_test.go:158` — elseif: `else { if` → `else if`
- [ ] `internal/todo/ui/tree.go:62` — paramTypeCombine: `matchedIssues, allIssues []*issue.Issue`
- [ ] `internal/todo/ui/tree.go:122` — paramTypeCombine: `issueByID, needed map[string]*issue.Issue`
- [ ] `internal/update/update.go:167` — nestingReduce: invert if, use continue

## unparam (16)
- [ ] `cmd/todo.go:41` — `initTodoCore`: cmd is unused
- [ ] `cmd/todo_query_test.go:39` — `createQueryTestIssue`: result 0 never used
- [ ] `cmd/todo_update.go:115` — `buildUpdateInput`: existingTags is unused
- [ ] `cmd/zed.go:37` — `resolveExt`: result 1 (error) always nil
- [ ] `internal/commit/commit_test.go:493` — `setupGitRepoWithRemote`: result dir never used
- [ ] `internal/todo/core/core_test.go:1299` — `writeTestIssueFile`: id is unused
- [ ] `internal/todo/core/watcher.go:162` — `unwatchLocked`: result 0 (error) always nil
- [ ] `internal/todo/graph/schema.resolvers_test.go:35` — `createTestIssue`: result 0 never used
- [ ] `internal/todo/integration/clickup/sync.go:687` — `syncRelationships`: result always nil
- [ ] `internal/todo/integration/clickup_adapter.go:409` — `validStatusList`: cfg is unused
- [ ] `internal/todo/integration/github_adapter_test.go:19` — `mustDetectGitHub`: owner always "o"
- [ ] `internal/todo/tui/list.go:44` — `treeAwareFilter`: deepSearch is unused
- [ ] `internal/todo/tui/prioritypicker.go:70` — `newPriorityPickerModel`: cfg is unused
- [ ] `internal/todo/tui/statuspicker.go:70` — `newStatusPickerModel`: cfg is unused
- [ ] `internal/todo/tui/typepicker.go:69` — `newTypePickerModel`: cfg is unused
- [ ] `internal/update/cite.go:137` — `parseSkill`: result 1 (error) always nil

## nilerr (9)
- [ ] `cmd/check.go:203` — error is not nil but returns nil (non-fatal)
- [ ] `internal/cite/add.go:120` — error is not nil but returns nil
- [ ] `internal/todo/config/config.go:253` — error is not nil but returns nil
- [ ] `internal/todo/core/watcher.go:136` — error is not nil but returns nil
- [ ] `internal/todo/core/watcher.go:276` — error is not nil but returns nil
- [ ] `internal/todo/core/watcher.go:297` — error is not nil but returns nil
- [ ] `internal/todo/core/watcher.go:313` — error is not nil but returns nil
- [ ] `internal/todo/integration/github_adapter.go:371` — error is not nil but returns nil
- [ ] `internal/update/commit.go:38` — error is not nil but returns nil

## staticcheck (3)
- [ ] `internal/todo/issue/issue_test.go:1536` — QF1001: apply De Morgan's law
- [ ] `internal/todo/tui/app_update_test.go:1345` — SA9003: empty branch
- [ ] `internal/todo/tui/app_update_test.go:1912` — SA1012: nil context → `context.TODO()`

## unused (3)
- [ ] `internal/todo/graph/filters_extended_test.go:675` — `setupTestResolverInDir` unused
- [ ] `internal/todo/tui/app_update_test.go:1716` — type `indexer` unused
- [ ] `internal/todo/tui/list.go:131` — `newItemDelegate` unused

## ineffassign (1)
- [ ] `cmd/check.go:163` — ineffectual assignment to `level`

## godoclint (1)
- [ ] `internal/todo/ui/styles.go:84` — godoc should start with `TagBadge`
