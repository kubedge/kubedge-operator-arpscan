# Tasks — buildx-multiarch-image

- [x] Confirm buildx + a builder are available (colima's docker on K8S01 provides buildx).
- [x] `docker-buildx` Makefile target present: `docker buildx build --platform linux/arm64,linux/amd64 -f build/Dockerfile -t <img> -t <repo>:latest --push`. `build/Dockerfile` cross-compiles per `TARGETOS/TARGETARCH` (fixes the old per-arch prebuilt-binary bug).
- [ ] `docker buildx imagetools inspect <image>:<tag>` → manifest list — deferred: needs a registry push (`make docker-buildx`), gated on `kubedge1` DockerHub creds.
- [x] Retired the arch-suffixed `docker-build-v1`/`docker-push-v1` targets and image names.
- [x] Image runs on arm64: smoke built the operator image via buildx and ran it on an arm64 kind node (operator pod 1/1 Ready).
