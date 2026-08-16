# Tasks — standardize-codegen

- [x] `controller-gen` at v0.21.0 (`go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.21.0`); `controller-gen --version` → v0.21.0.
- [x] `generate` target carries no removed `crd:trivialVersions=true` flag (graft's Makefile is a clean rewrite).
- [x] No hard-coded `/usr/local/kubebuilder/bin` PATH in the graft Makefile.
- [x] `make generate` regenerates the CRD under `chart/crds/` (NOT `chart/templates/` — the ~3.2MB embedded PodTemplateSpec schemas would blow Helm's 1MB release-Secret limit). No local deepcopy step: base owns DeepCopy (this repo is a base consumer). Removed the stale `chart/templates/…_arpscans.yaml`.
- [x] `go build ./...` green after regeneration; CRD installs and the sample CR validates against it (smoke).
