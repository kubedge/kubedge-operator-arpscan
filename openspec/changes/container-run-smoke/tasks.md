# Tasks — container-run-smoke

- [x] DECISION: kind on colima's docker (K8S01) — `bin/smoke.sh` / `make smoke`.
- [x] `make generate` → CRD applied via the Helm chart's `chart/crds/`; registers in-cluster (sample CR validates against it).
- [x] Build the consumer operator image against base (buildx, arm64) and deploy it (`bin/smoke.sh up`).
- [x] Apply the sample CR (`examples/example-arpscan.yaml`); operator pod starts and the CR reaches `status.satisfied=true` (actualState == targetState `deployed`) and the operator creates the `arpscan` Deployment.
- [x] Teardown: `bin/smoke.sh down`/`nuke` removes the CR + Helm release (owner-ref GC removes the rendered `arpscan` Deployment). No `make undeploy` target — teardown lives in the smoke script.

> NOTE: the rendered scanner workload pod (`hack4easy/arpscan-amd64`, amd64-only, `/bin/arpscan eth0`, privileged) does not reach Ready on an arm64 kind node — a property of that external image, not this operator. The smoke asserts operator-up + reconcile-to-satisfied + workload-created, not scanner-pod readiness.
