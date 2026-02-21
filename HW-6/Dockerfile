FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./
COPY model/ ./model/
COPY store/ ./store/
COPY seeddata/ ./seeddata/
COPY generator/ ./generator/
COPY search/ ./search/
COPY handler/ ./handler/

RUN go build -o product-search .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/product-search .

EXPOSE 8080

CMD ["./product-search"]