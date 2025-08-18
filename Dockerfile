# Build the SPIFFE CSI Driver binary
FROM --platform=${BUILDPLATFORM} golang:1.24.6-alpine AS base
WORKDIR /code
RUN apk --no-cache --update add make
COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

# xx is a helper for cross-compilation
# when bumping to a new version analyze the new version for security issues
# then use crane to lookup the digest of that version so we are immutable
# crane digest tonistiigi/xx:1.1.2
FROM --platform=${BUILDPLATFORM} tonistiigi/xx@sha256:9dde7edeb9e4a957ce78be9f8c0fbabe0129bf5126933cd3574888f443731cda AS xx

FROM --platform=${BUILDPLATFORM} base AS builder
ARG TARGETPLATFORM
ARG TARGETARCH
ENV CGO_ENABLED=0
COPY --link --from=xx / /
RUN xx-go --wrap
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make build

# Build a scratch image with just the SPIFFE CSI driver binary
FROM scratch AS spiffe-csi-driver
WORKDIR /
ENTRYPOINT ["/spiffe-csi-driver"]
CMD []
COPY --link --from=builder /code/bin/spiffe-csi-driver /spiffe-csi-driver
