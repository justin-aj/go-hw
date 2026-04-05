# ---- build stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN go build -o kv-node .

# ---- runtime stage ----
FROM alpine:3.21
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=builder /app/kv-node .
EXPOSE 8080
CMD ["./kv-node"]
