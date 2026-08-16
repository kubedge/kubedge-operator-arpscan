# Tasks — adopt-go-ci

- [x] From the meta session, `/alemax:update-skills` staged the class-M set (incl. `ci.yml`) — delivered as `meta-broadcast/deliver-class-m-templates-ciyml--hygiene`.
- [x] In this repo, `/alemax:complete-update` applied the broadcast onto `main` (Phase 1).
- [x] `.github/workflows/ci.yml` present; its Go jobs gate on `go.mod` (graft's fixed ci.yml supersedes the stale broadcast copy).
- [ ] Trial push → confirm `go-build`/`go-vet`/`go-test`/`golangci-lint` green — happens on the PR push (CI run).
- [x] Rest of class-M landed: `.editorconfig`, `.gitattributes`, `.github/*`, `dependabot.yml`, `.pre-commit-config.yaml`, `bin/set-secret.sh`.
