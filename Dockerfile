# Build the SPIFFE CSI Driver binary
FROM golang:1.19.1-alpine AS builder
ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_DIRTY
RUN apk add make
WORKDIR /code
COPY go.mod /code/go.mod
COPY go.sum /code/go.sum
RUN go mod download
ADD . /code
RUN CGO_ENABLED=0 go test ./...
RUN CGO_ENABLED=0 make GIT_TAG="${GIT_TAG}" GIT_COMMIT="${GIT_COMMIT}" GIT_DIRTY="${GIT_DIRTY}" build

# Build a scratch image with just the SPIFFE CSI driver binary
FROM scratch AS spiffe-csi-driver
COPY --from=builder /code/bin/spiffe-csi-driver /spiffe-csi-driver
WORKDIR /
ENTRYPOINT ["/spiffe-csi-driver"]
CMD []
