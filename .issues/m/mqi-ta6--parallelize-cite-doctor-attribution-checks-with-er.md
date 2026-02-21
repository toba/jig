---
# mqi-ta6
title: Parallelize cite doctor attribution checks with errgroup
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:04Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
---

## Description
`cite/doctor.go:68-76` runs `checkAttribution` sequentially for each cited source. Each makes independent HTTP calls via `gh api`.

## TODO
- [x] Run `checkAttribution` calls concurrently with bounded `errgroup.Group`
- [x] Collect boolean results safely
- [x] Run tests
