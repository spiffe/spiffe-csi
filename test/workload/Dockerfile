FROM golang:1.24.9-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /workload

CMD [ "/workload" ]
