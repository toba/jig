---
# sqz-7ri
title: brew doctor should accept goreleaser v2 format (singular) in addition to formats (list)
status: completed
type: bug
priority: normal
created_at: 2026-03-05T19:40:52Z
updated_at: 2026-03-05T19:47:38Z
sync:
    github:
        issue_number: "78"
        synced_at: "2026-03-05T19:49:30Z"
---

`jig brew doctor` only checks the `formats` (plural list) field on goreleaser archives. Goreleaser v2 also supports `format` (singular string) which is the more common syntax:

```yaml
archives:
  - format: tar.gz
```

The doctor should accept both `format: tar.gz` and `formats: [tar.gz]`.

## Summary of Changes

Added support for goreleaser v2 singular `format` field in both brew and scoop doctor checks:

1. **brew/doctor.go**: Added `Format string` field to the goreleaser archive struct. The tar.gz check now accepts both `format: tar.gz` (v2) and `formats: [tar.gz]` (v1).
2. **scoop/doctor.go**: Same change — added `Format string` field and the zip check now accepts both `format: zip` and `formats: [zip]`.
