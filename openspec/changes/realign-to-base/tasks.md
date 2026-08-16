# Tasks — realign-to-base

- [x] `go get github.com/kubedge/kubedge-operator-base@v0.1.36-kubedge.20260815`
- [x] `go mod tidy` (confirmed k8s.io/* → v0.36.3, controller-runtime → v0.24.1, go 1.26)
- [x] `go build ./...` → fixed controller-runtime Watch API (`source.Kind` now takes the object + handler; wrapped as `client.Object`)
- [x] `go vet ./...` green; `go test ./...` green (no test files yet — see test-coverage-uplift)
- [ ] `golangci-lint run …` — deferred to CI (the class-M `ci.yml` runs it on push)
- [x] Operator builds its binary and runs: smoke deploy on kind → operator pod 1/1 Ready, Arpscan `satisfied=true`
