FROM --platform=linux/amd64 golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Garantir build para AMD64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker ./cmd/main.go

FROM --platform=linux/amd64 alpine:latest

WORKDIR /app

COPY --from=builder /app/worker .

EXPOSE 8080

ENTRYPOINT ["./worker"]