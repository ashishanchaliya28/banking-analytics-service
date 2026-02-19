FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-service ./cmd/main.go

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata && adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /analytics-service .
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["./analytics-service"]
