---
# b8s-42z
title: Investigate tree-sitter AST parsing for shell commands
status: draft
type: task
created_at: 2026-02-21T19:50:20Z
updated_at: 2026-02-21T19:50:20Z
---

The guardrails repo uses tree-sitter-bash for AST-based parsing, which catches quote-splitting evasion like r"m" -rf / by reassembling quoted fragments. nope's ShellTokenize handles quoting for token classification but doesn't reassemble fragments to detect the underlying command.

Significant dependency cost (tree-sitter). The regex-fallback pattern (if AST fails, still apply regex) is a good safety pattern regardless.

Reference: https://github.com/leegonzales/claude-guardrails (tree-sitter-bash usage)

- [ ] Evaluate Go tree-sitter bindings maturity
- [ ] Assess whether quote-splitting evasion is a real risk given nope's tokenizer
- [ ] Consider lighter alternatives (token reassembly without full AST)
