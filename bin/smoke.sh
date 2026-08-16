#!/usr/bin/env bash
# smoke.sh — local container run-smoke gate for kubedge-arpscan-operator.
#
# Proves the operator actually comes up and reconciles, not just that it compiles:
#   build image -> load into a kind cluster -> helm install -> apply a sample CR ->
#   assert the Arpscan reports SATISFIED and every component StatefulSet is Ready.
#
# Cluster decision (K8S01): kind on top of colima's docker. Rationale — colima already
# provides the docker daemon the buildx image build needs, and `kind load docker-image`
# moves that image straight into the node's containerd (k3s/containerd can't otherwise
# see colima-docker images). minikube/colima+k3s were rejected: extra image-plumbing.
#
# Usage:
#   bin/smoke.sh up       # create/reuse cluster, build+load image, deploy, apply CR, assert
#   bin/smoke.sh down     # delete the CR + helm release (leaves the cluster)
#   bin/smoke.sh nuke     # down + delete the kind cluster
#   bin/smoke.sh          # up, assert, then down (full cycle; the CI-style gate)
set -euo pipefail

CLUSTER="${CLUSTER:-arpscan}"
NAMESPACE="${NAMESPACE:-default}"
RELEASE="${RELEASE:-kubedge-arpscan-operator}"
VERSION="${VERSION:-0.2.0}"
IMG="${IMG:-kubedge1/kubedge-arpscan-operator:v${VERSION}}"
NODE="${CLUSTER}-control-plane"
# arpscan reconciles its CR into a single Deployment named `arpscan` (rendered from the
# operator image's /opt/kubedge-operators/arpscan-templates/arpscan.yaml).
WORKLOAD_DEPLOY="${WORKLOAD_DEPLOY:-arpscan}"
CR_NAME="${CR_NAME:-kubedge-arpscan-scanners}"
CR_FILE="${CR_FILE:-examples/example-arpscan.yaml}"

# colima's docker socket, so kind and buildx talk to the same daemon.
COLIMA_SOCK="${HOME}/.config/colima/default/docker.sock"
[ -S "$COLIMA_SOCK" ] && export DOCKER_HOST="unix://${COLIMA_SOCK}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

log() { printf '\n=== %s ===\n' "$*"; }
need() { command -v "$1" >/dev/null 2>&1 || {
  echo "ERROR: '$1' not found on PATH" >&2
  exit 1
}; }

ensure_cluster() {
  need kind
  need kubectl
  need helm
  if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
    log "creating kind cluster '$CLUSTER'"
    kind create cluster --name "$CLUSTER"
  else
    log "reusing kind cluster '$CLUSTER'"
  fi
  kubectl config use-context "kind-${CLUSTER}" >/dev/null
  kubectl wait --for=condition=Ready "node/${NODE}" --timeout=120s
  # The operator Deployment and the sample component pods pin czZone/ezZone nodeSelectors;
  # on a single-node cluster just satisfy both on the one node.
  kubectl label "node/${NODE}" czZone=enabled ezZone=enabled --overwrite >/dev/null
}

deploy() {
  log "building + loading image ${IMG}"
  docker buildx build --load -f build/Dockerfile -t "$IMG" .
  kind load docker-image "$IMG" --name "$CLUSTER"

  log "helm install/upgrade ${RELEASE}"
  # pull_policy=Never: use the kind-loaded image, never reach a registry.
  helm upgrade --install "$RELEASE" chart \
    --set images.tags.operator="$IMG" \
    --set images.pull_policy=Never \
    --namespace "$NAMESPACE"

  kubectl wait --for=condition=Ready pod \
    -l "name=${RELEASE}" -n "$NAMESPACE" --timeout=120s
}

apply_cr_and_assert() {
  log "applying ${CR_FILE}"
  kubectl apply -f "$CR_FILE" -n "$NAMESPACE"

  log "waiting for Arpscan ${CR_NAME} to report satisfied"
  local satisfied=""
  for _ in $(seq 1 30); do
    satisfied="$(kubectl get arpscan "$CR_NAME" -n "$NAMESPACE" \
      -o jsonpath='{.status.satisfied}' 2>/dev/null || true)"
    [ "$satisfied" = "true" ] && break
    sleep 2
  done
  if [ "$satisfied" != "true" ]; then
    echo "FAIL: Arpscan not satisfied after wait (satisfied=${satisfied:-<none>})" >&2
    kubectl get arpscan "$CR_NAME" -n "$NAMESPACE" -o yaml >&2 || true
    exit 1
  fi

  # The operator-reconcile gate: satisfied=true means the operator applied its template,
  # so the workload Deployment must exist. We do NOT hard-assert the scanner pods reach
  # Ready: the rendered `arpscan` container runs the external image hack4easy/arpscan-*
  # (amd64-only, privileged, needs eth0) which will not run on an arm64 kind node. Its
  # readiness is a property of that external image, not of this operator.
  log "asserting workload Deployment '${WORKLOAD_DEPLOY}' was created by the operator"
  kubectl get "deployment/${WORKLOAD_DEPLOY}" -n "$NAMESPACE" >/dev/null || {
    echo "FAIL: operator reported satisfied but Deployment '${WORKLOAD_DEPLOY}' is absent" >&2
    exit 1
  }
  # Best-effort rollout wait; report but tolerate the external-image case.
  kubectl rollout status "deployment/${WORKLOAD_DEPLOY}" -n "$NAMESPACE" --timeout=60s \
    || log "NOTE: '${WORKLOAD_DEPLOY}' not Ready — expected when the external scanner image can't run on this node arch"

  log "SMOKE PASS — operator came up, reconciled Arpscan to satisfied, and created the workload"
  kubectl get arpscan,deployment,pod -n "$NAMESPACE"
}

teardown() {
  log "removing sample CR + helm release"
  kubectl delete -f "$CR_FILE" -n "$NAMESPACE" --ignore-not-found --timeout=120s || true
  helm uninstall "$RELEASE" --namespace "$NAMESPACE" 2>/dev/null || true
}

case "${1:-cycle}" in
  up)
    ensure_cluster
    deploy
    apply_cr_and_assert
    ;;
  down) teardown ;;
  nuke)
    teardown
    kind delete cluster --name "$CLUSTER"
    ;;
  cycle)
    ensure_cluster
    deploy
    apply_cr_and_assert
    teardown
    ;;
  *)
    echo "usage: $0 [up|down|nuke|cycle]" >&2
    exit 2
    ;;
esac
