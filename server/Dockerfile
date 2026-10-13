FROM golang:1.27-alpine AS build
WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /mall . && CGO_ENABLED=0 go build -trimpath -o /seed ./cmd/seed
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S mall && adduser -S mall -G mall
COPY --from=build /mall /usr/local/bin/mall
COPY --from=build /seed /usr/local/bin/seed
USER mall
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=4s --retries=5 CMD wget -qO- http://127.0.0.1:8080/health || exit 1
CMD ["mall"]
