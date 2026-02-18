# Multistage build para o Video Processor Worker

# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Instala dependências de build
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copia arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Build do binário com otimizações
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o worker \
    cmd/worker/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Instala FFmpeg e certificados SSL
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata

# Cria usuário não-root para segurança
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Cria diretórios necessários (temp será criado pelo app)
RUN mkdir -p outputs uploads && \
    chown -R appuser:appgroup . && \
    chmod -R 755 .

# Copia o binário do stage de build
COPY --from=builder --chown=appuser:appgroup /build/worker .

# Garante permissões executáveis do binário
RUN chmod +x ./worker

# Muda para usuário não-root
USER appuser

# Expõe porta (se necessário no futuro)
# EXPOSE 8080

# Healthcheck (opcional)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ps aux | grep -q '[w]orker' || exit 1

# Comando para executar o worker
CMD ["./worker"]
