---
# krz-eaz
title: Fix all golangci-lint issues
status: completed
type: task
priority: normal
created_at: 2026-02-24T17:35:30Z
updated_at: 2026-03-21T17:59:48Z
sync:
    github:
        issue_number: "67"
        synced_at: "2026-03-21T18:02:31Z"
---

Work through all 122 remaining golangci-lint issues that require manual intervention. Run `./scripts/lint.sh` to verify progress.

## errcheck (50)

### Production code
- [x] `internal/brew/init.go:212` — `defer os.RemoveAll(tmp)`
- [x] `internal/brew/sha.go:68` — `defer os.RemoveAll(tmp)`
- [x] `internal/cite/add.go:97` — `defer os.RemoveAll(tmpDir)`
- [x] `internal/display/display.go:55` — `fmt.Fprintln(w, sepStyle.Render("---"))`
- [x] `internal/display/display.go:56` — `fmt.Fprintln(w)`
- [x] `internal/display/display.go:62` — `fmt.Fprintln(w, header)`
- [x] `internal/display/display.go:82` — `fmt.Fprintf(w, ...)`
- [x] `internal/nope/debug.go:44` — `d.f.Write(data)`
- [x] `internal/nope/debug.go:45` — `d.f.WriteString("\n")`
- [x] `internal/nope/debug.go:53` — `d.f.Close()`
- [x] `internal/testutil/chdir.go:17` — `os.Chdir(orig)` in cleanup
- [x] `internal/todo/core/core.go:92` — `fmt.Fprintf(c.warnWriter, ...)`
- [x] `internal/todo/core/core.go:145` — `c.searchIndex.Close()`
- [x] `internal/todo/core/core.go:162` — `defer f.Close()`
- [x] `internal/todo/core/watcher.go:128` — `watcher.Close()`
- [x] `internal/todo/core/watcher.go:184` — `defer watcher.Close()`
- [x] `internal/todo/integration/syncutil/images.go:92` — `defer f.Close()`

### Test code
- [x] `internal/cite/doctor_test.go:82` — `os.WriteFile("LICENSE", ...)`
- [x] `internal/cite/doctor_test.go:83` — `os.WriteFile("NOTICE", ...)`
- [x] `internal/cite/doctor_test.go:119` — `os.WriteFile("LICENSE", ...)`
- [x] `internal/nope/guard_test.go:72` — `w.Close()`
- [x] `internal/todo/core/core_test.go:399` — `os.Mkdir(..., "subdir")`
- [x] `internal/todo/core/core_test.go:649` — `defer core.Unwatch()`
- [x] `internal/todo/core/core_test.go:697` — `defer core.Unwatch()`
- [x] `internal/todo/core/core_test.go:740` — `defer core.Unwatch()`
- [x] `internal/todo/core/core_test.go:884` — `os.Remove(...)`
- [x] `internal/todo/core/search_test.go:12` — `defer core.Close()`
- [x] `internal/todo/core/search_test.go:40` — `defer core.Close()`
- [x] `internal/todo/core/search_test.go:65` — `defer core.Close()`
- [x] `internal/todo/core/watcher_test.go:31` — `defer watcher.Close()`
- [x] `internal/todo/integration/github/images_test.go:19` — `json.NewEncoder(w).Encode(...)`
- [x] `internal/todo/integration/github/images_test.go:40` — `json.NewEncoder(w).Encode(...)`
- [x] `internal/todo/integration/github/images_test.go:64` — `json.NewEncoder(w).Encode(...)`
- [x] `internal/todo/integration/github/sync_test.go:979` — `fmt.Sscanf(...)`
- [x] `internal/todo/integration/github/sync_test.go:1052` — `fmt.Sscanf(...)`
- [x] `internal/todo/integration/github/sync_test.go:1055` — `fmt.Sscanf(...)`
- [x] `internal/todo/issue/issue_test.go:2209` — `json.Unmarshal(data1, ...)`
- [x] `internal/todo/issue/issue_test.go:2210` — `json.Unmarshal(data2, ...)`
- [x] `internal/todo/output/output_test.go:26` — `w.Close()`
- [x] `internal/todo/output/output_test.go:31` — `r.Close()`
- [x] `internal/todo/search/index_test.go:17` — `idx.Close()`
- [x] `internal/todo/search/index_test.go:27` — `defer idx.Close()`
- [x] `internal/update/commit_test.go:64` — `os.Chdir(orig)` in cleanup
- [x] `internal/update/commit_test.go:65` — `os.Chdir(tmp)`
- [x] `internal/zed/doctor_test.go:197` — `os.Setenv("PATH", ...)`
- [x] `internal/zed/doctor_test.go:198` — `defer os.Setenv("PATH", ...)`
- [x] `internal/zed/doctor_test.go:207` — `os.MkdirAll(".github/workflows", ...)`
- [x] `internal/zed/doctor_test.go:248` — `os.Setenv("PATH", ...)`
- [x] `internal/zed/doctor_test.go:290` — `os.MkdirAll(".github/workflows", ...)`
- [x] `internal/zed/doctor_test.go:332` — `os.MkdirAll(".github/workflows", ...)`

## gosec (21)
- [x] `internal/commit/commit.go:82` — G204: subprocess with tainted input (`git log tag+..HEAD`)
- [x] `internal/commit/commit_test.go:84` — G204: subprocess with variable
- [x] `internal/commit/commit_test.go:256` — G204: subprocess with variable
- [x] `internal/companion/gh.go:29` — G204: subprocess with variable (`gh` args)
- [x] `internal/companion/gh.go:47` — G204: subprocess with variable (`gh release list`)
- [x] `internal/config/companions.go:49` — G306: WriteFile 0o644 perms
- [x] `internal/config/config.go:43` — G304: file inclusion via variable
- [x] `internal/config/config.go:101` — G306: WriteFile 0o644 perms
- [x] `internal/config/config.go:208` — G306: WriteFile 0o644 perms
- [x] `internal/config/config_test.go:110` — G304: file inclusion via variable
- [x] `internal/nope/config.go:24` — G101: `BuiltinCredentialRead` const name
- [x] `internal/testutil/readfile.go:10` — G304: file inclusion via variable
- [x] `internal/todo/config/config_test.go:522` — G301: MkdirAll 0755 perms
- [x] `internal/todo/core/core.go:422` — G104: unhandled `h.Write(content)`
- [x] `internal/todo/integration/syncutil/retry.go:70` — G404: weak random (`rand.Int64N`)
- [x] `internal/todo/integration/syncutil/retry.go:90` — G704: SSRF via taint analysis
- [x] `internal/todo/issue/issue.go:346` — G104: unhandled `h.Write(content)`
- [x] `internal/update/commit.go:101` — G104: unhandled `os.Remove(dir)`
- [x] `internal/update/commit_test.go:69` — G301: MkdirAll 0o755 perms
- [x] `internal/update/commit_test.go:80` — G301: MkdirAll 0o755 perms
- [x] `pkg/client/client_test.go:15` — G204: subprocess with variable

## gocritic (18)
- [x] `cmd/init.go:45` — emptyStringTest: `len(content) > 0` → `content != ""`
- [x] `cmd/todo_content.go:60` — paramTypeCombine: `code string, format string` → `code, format string`
- [x] `internal/cite/add.go:259` — paramTypeCombine: `slice []string, files []string`
- [x] `internal/nope/guard.go:71` — paramTypeCombine: `toolName string, input string`
- [x] `internal/nope/init.go:132` — emptyStringTest: `len(content) > 0`
- [x] `internal/todo/core/links.go:142` — appendAssign: append result not assigned to same slice
- [x] `internal/todo/core/links.go:241` — appendAssign: append result not assigned to same slice
- [x] `internal/todo/core/watcher.go:113` — deprecatedComment: needs dedicated paragraph
- [x] `internal/todo/graph/schema.resolvers.go:359` — paramTypeCombine: `id string, name string`
- [x] `internal/todo/graph/schema.resolvers.go:378` — paramTypeCombine: `id string, name string`
- [x] `internal/todo/integration/clickup/config.go:142` — singleCaseSwitch: rewrite to if
- [x] `internal/todo/integration/github/sync.go:475` — appendCombine: combine 2 appends
- [x] `internal/todo/output/output.go:94` — paramTypeCombine: `code string, message string`
- [x] `internal/todo/tui/tui.go:436` — appendAssign: append result not assigned to same slice
- [x] `internal/todo/tui/tui_test.go:158` — elseif: `else { if` → `else if`
- [x] `internal/todo/ui/tree.go:62` — paramTypeCombine: `matchedIssues, allIssues []*issue.Issue`
- [x] `internal/todo/ui/tree.go:122` — paramTypeCombine: `issueByID, needed map[string]*issue.Issue`
- [x] `internal/update/update.go:167` — nestingReduce: invert if, use continue

## unparam (16)
- [x] `cmd/todo.go:41` — `initTodoCore`: cmd is unused
- [x] `cmd/todo_query_test.go:39` — `createQueryTestIssue`: result 0 never used
- [x] `cmd/todo_update.go:115` — `buildUpdateInput`: existingTags is unused
- [x] `cmd/zed.go:37` — `resolveExt`: result 1 (error) always nil
- [x] `internal/commit/commit_test.go:493` — `setupGitRepoWithRemote`: result dir never used
- [x] `internal/todo/core/core_test.go:1299` — `writeTestIssueFile`: id is unused
- [x] `internal/todo/core/watcher.go:162` — `unwatchLocked`: result 0 (error) always nil
- [x] `internal/todo/graph/schema.resolvers_test.go:35` — `createTestIssue`: result 0 never used
- [x] `internal/todo/integration/clickup/sync.go:687` — `syncRelationships`: result always nil
- [x] `internal/todo/integration/clickup_adapter.go:409` — `validStatusList`: cfg is unused
- [x] `internal/todo/integration/github_adapter_test.go:19` — `mustDetectGitHub`: owner always "o"
- [x] `internal/todo/tui/list.go:44` — `treeAwareFilter`: deepSearch is unused
- [x] `internal/todo/tui/prioritypicker.go:70` — `newPriorityPickerModel`: cfg is unused
- [x] `internal/todo/tui/statuspicker.go:70` — `newStatusPickerModel`: cfg is unused
- [x] `internal/todo/tui/typepicker.go:69` — `newTypePickerModel`: cfg is unused
- [x] `internal/update/cite.go:137` — `parseSkill`: result 1 (error) always nil

## nilerr (9)
- [x] `cmd/check.go:203` — error is not nil but returns nil (non-fatal)
- [x] `internal/cite/add.go:120` — error is not nil but returns nil
- [x] `internal/todo/config/config.go:253` — error is not nil but returns nil
- [x] `internal/todo/core/watcher.go:136` — error is not nil but returns nil
- [x] `internal/todo/core/watcher.go:276` — error is not nil but returns nil
- [x] `internal/todo/core/watcher.go:297` — error is not nil but returns nil
- [x] `internal/todo/core/watcher.go:313` — error is not nil but returns nil
- [x] `internal/todo/integration/github_adapter.go:371` — error is not nil but returns nil
- [x] `internal/update/commit.go:38` — error is not nil but returns nil

## staticcheck (3)
- [x] `internal/todo/issue/issue_test.go:1536` — QF1001: apply De Morgan's law
- [x] `internal/todo/tui/app_update_test.go:1345` — SA9003: empty branch
- [x] `internal/todo/tui/app_update_test.go:1912` — SA1012: nil context → `context.TODO()`

## unused (3)
- [x] `internal/todo/graph/filters_extended_test.go:675` — `setupTestResolverInDir` unused
- [x] `internal/todo/tui/app_update_test.go:1716` — type `indexer` unused
- [x] `internal/todo/tui/list.go:131` — `newItemDelegate` unused

## ineffassign (1)
- [x] `cmd/check.go:163` — ineffectual assignment to `level`

## godoclint (1)
- [x] `internal/todo/ui/styles.go:84` — godoc should start with `TagBadge`


## Summary of Changes

All 122 golangci-lint issues resolved. `scripts/lint.sh` reports 0 issues.
