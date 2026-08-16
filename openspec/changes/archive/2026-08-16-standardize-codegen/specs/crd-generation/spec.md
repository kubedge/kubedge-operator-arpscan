## ADDED Requirements

### Requirement: `make generate` regenerates CRDs and deepcopy with a modern controller-gen

`make generate` SHALL run against a pinned modern controller-gen (v0.21.0, installed via
`go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.21.0`), SHALL NOT pass the
removed `crd:trivialVersions=true` flag, and SHALL NOT depend on a hard-coded
`/usr/local/kubebuilder/bin` PATH. It reads the CRD types from the `kubedge-operator-base`
module (this repo is a base *consumer*, so base owns the DeepCopy code and there is no
local deepcopy step) and writes the CRD manifest to `chart/crds/` — not `chart/templates/`,
because the `arpscans` CRD embeds ~3.2MB of PodTemplateSpec schemas and would exceed Helm's
1MB release-Secret limit if rendered as a template.

#### Scenario: generate succeeds on a modern toolchain
- **WHEN** `make generate` runs with controller-gen v0.21.0 on PATH
- **THEN** it completes without the `trivialVersions` error, writes `chart/crds/kubedgeoperators.kubedge.cloud_arpscans.yaml`, and `go build ./...` stays green
