FROM golang:1.25.3 AS builder
WORKDIR /app
COPY go.mod go.sum /app/

RUN go mod download

COPY ddflare.go /app/ddflare.go
COPY cli /app/cli
COPY pkg /app/pkg

ENV CGO_ENABLED=0
ARG VERSION=v0.0.0
RUN go build \
    -ldflags "-w -s \
    -X github.com/ddflare/ddflare/pkg/version.Version=$VERSION" \
    -o /ddflare \
    /app/cli

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /ddflare .

ENTRYPOINT ["/ddflare"]
