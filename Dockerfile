FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum /app

RUN go mod download

COPY main.go /app
COPY cmd /app/cmd
COPY pkg /app/pkg

ENV CGO_ENABLED=0
ARG VERSION=v0.0.0
RUN go build \
    -ldflags "-w -s \
    -X github.com/fgiudici/ddflare/pkg/version.Version=$VERSION" \
    -o /ddflare

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /ddflare .

ENTRYPOINT ["/ddflare"]