FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /metrics-batch-collector ./cmd/app

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /metrics-batch-collector /app/metrics-batch-collector

EXPOSE 8080

ENTRYPOINT ["/app/metrics-batch-collector"]
