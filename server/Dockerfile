# Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Run
FROM scratch
COPY --from=builder /app/server /server
EXPOSE 4000
ENTRYPOINT ["/server"]
CMD ["-port", "4000", "-env", "production", "-redis-addr", "redis:6379"]
