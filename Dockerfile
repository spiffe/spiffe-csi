# Build the SPIFFE CSI Driver binary
FROM golang:1.19.3-alpine AS builder
ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_DIRTY
WORKDIR /code
RUN apk --no-cache --update add make
COPY go.* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN make test
RUN make GIT_TAG="${GIT_TAG}" GIT_COMMIT="${GIT_COMMIT}" GIT_DIRTY="${GIT_DIRTY}" build

# Build a scratch image with just the SPIFFE CSI driver binary
FROM scratch AS spiffe-csi-driver
WORKDIR /
ENTRYPOINT ["/spiffe-csi-driver"]
CMD []
COPY --from=builder /code/bin/spiffe-csi-driver /spiffe-csi-driver
