# syntax=docker/dockerfile:1.7
#
# Multi-stage build for the Go backend API.
#
# Stage 1 (builder): compiles a static, fully-stripped binary.
# Stage 2 (runtime): distroless image with the binary only — no shell,
# no package manager, no Go toolchain. Smallest possible attack surface.

# -----------------------------------------------------------------------------
# Stage 1: builder
# -----------------------------------------------------------------------------
FROM golang:1.25.1-alpine AS builder

# CGO_ENABLED=0 produces a statically-linked binary that works in any
# minimal base image (no glibc/musl dependency on the runtime side).
# GOOS/GOARCH are the defaults but pinning them makes the intent explicit
# and protects against future `--platform` builds doing something unexpected.
# -trimpath strips absolute paths from the binary for reproducible builds.
# -ldflags "-s -w" strips the symbol table and DWARF debug info.
# -buildvcs=false skips embedding VCS info (otherwise `git` would be required
# in the build context, breaking `docker build` from a tarball).
WORKDIR /src

# Copy module files first and download dependencies as a separate cached layer.
# If go.mod/go.sum don't change, this layer is reused on rebuilds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Now copy the rest of the source and build.
COPY . .

# Run tests inside the build stage. If any unit test fails, the image
# build fails — guarantees we never ship a broken binary.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go test ./... -count=1

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -buildvcs=false \
      -ldflags="-s -w" \
      -o /out/server \
      ./cmd/server

# -----------------------------------------------------------------------------
# Stage 2: runtime
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

# Copy CA certificates from the builder so any outbound HTTPS the app makes
# (e.g. external APIs) can verify TLS without bundling a CA store manually.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary. --chown ensures the nonroot user owns it.
COPY --from=builder --chown=nonroot:nonroot /out/server /usr/local/bin/server

# Run as the built-in `nonroot` user (UID 65532). Combined with distroless
# (no shell), this means even if the app is exploited, the attacker has
# no shell and no privilege escalation path inside the container.
USER nonroot:nonroot

# Document the port; --network=host or compose port mapping uses this.
EXPOSE 8080

# Healthcheck is intentionally NOT in the Dockerfile. Why: distroless has
# no shell, so a HEALTHCHECK CMD would have to be an absolute binary path
# to something that can hit /health. The right place to define it is the
# orchestrator (docker-compose, k8s readiness probe) which can run `wget`,
# `curl`, or a Go binary built for that purpose.
#
# Recommended compose healthcheck:
#   test: ["CMD", "/usr/local/bin/server", "-healthcheck"]
# Or an external check against http://localhost:8080/health.

ENTRYPOINT ["/usr/local/bin/server"]
