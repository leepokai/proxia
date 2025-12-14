FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/gateway .

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/gateway /app/gateway
USER app
EXPOSE 8080
CMD ["/app/gateway"]


