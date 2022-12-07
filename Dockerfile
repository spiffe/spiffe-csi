# Build the SPIFFE CSI Driver binary
FROM --platform=${BUILDPLATFORM} golang:1.19.3-alpine AS base
ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_DIRTY
WORKDIR /code
RUN apk --no-cache --update add make
COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

# xx is a helper for cross-compilation
FROM --platform=${BUILDPLATFORM} tonistiigi/xx:1.1.2 AS xx

FROM --platform=${BUILDPLATFORM} base as builder
ARG TARGETPLATFORM
ARG TARGETARCH
ENV CGO_ENABLED=0
COPY --link --from=xx / /
RUN xx-go --wrap
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    if [ "$TARGETARCH" = "arm64" ]; then CC=aarch64-alpine-linux-musl; fi && \
    make GIT_TAG="${GIT_TAG}" GIT_COMMIT="${GIT_COMMIT}" GIT_DIRTY="${GIT_DIRTY}" build

# Build a scratch image with just the SPIFFE CSI driver binary
FROM scratch AS spiffe-csi-driver
WORKDIR /
ENTRYPOINT ["/spiffe-csi-driver"]
CMD []
COPY --link --from=builder /code/bin/spiffe-csi-driver /spiffe-csi-driver
