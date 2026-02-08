.PHONY: help build run stop clean test docker-build docker-run docker-stop docker-clean docker-logs

# Variáveis
DOCKER_IMAGE_NAME = hackaton-soat-processor
DOCKER_IMAGE_TAG = latest
BINARY_NAME = worker

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Comandos Go locais
build: ## Compila o binário localmente
	@echo "🔨 Compilando binário..."
	go build -o $(BINARY_NAME) cmd/worker/main.go
	@echo "✅ Binário criado: $(BINARY_NAME)"

run: ## Executa o worker localmente
	@echo "🚀 Executando worker..."
	go run cmd/worker/main.go

test: ## Executa todos os testes
	@echo "🧪 Executando testes..."
	go test -v ./...

test-cover: ## Executa testes com cobertura
	@echo "🧪 Executando testes com cobertura..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Relatório de cobertura gerado: coverage.html"

clean: ## Remove binários e arquivos temporários
	@echo "🧹 Limpando arquivos..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf temp/
	rm -rf outputs/
	@echo "✅ Limpeza concluída"

# Comandos Docker
docker-build: ## Constrói a imagem Docker
	@echo "🐳 Construindo imagem Docker..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo "✅ Imagem construída: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

docker-run: ## Executa o container Docker
	@echo "🐳 Iniciando container..."
	docker-compose up -d
	@echo "✅ Container iniciado"

docker-stop: ## Para o container Docker
	@echo "🛑 Parando container..."
	docker-compose down
	@echo "✅ Container parado"

docker-logs: ## Mostra os logs do container
	docker-compose logs -f worker

docker-clean: ## Remove imagens e containers Docker
	@echo "🧹 Limpando Docker..."
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) 2>/dev/null || true
	@echo "✅ Limps3://teste-video-processor/processed/frames_f0f0dbb2-197c-4be0-b020-2caa043bba33.zipeza Docker concluída"

docker-shell: ## Abre um shell no container
	docker-compose exec worker /bin/sh

docker-rebuild: docker-clean docker-build ## Reconstrói a imagem do zero

# Comandos de desenvolvimento
dev: ## Inicia o ambiente de desenvolvimento
	@echo "🔧 Iniciando ambiente de desenvolvimento..."
	docker-compose up --build

deps: ## Baixa as dependências do Go
	@echo "📦 Baixando dependências..."
	go mod download
	go mod tidy
	@echo "✅ Dependências atualizadas"

fmt: ## Formata o código Go
	@echo "✨ Formatando código..."
	go fmt ./...
	@echo "✅ Código formatado"

lint: ## Executa o linter
	@echo "🔍 Executando linter..."
	golangci-lint run ./... || echo "⚠️  Instale golangci-lint: https://golangci-lint.run/usage/install/"

# Comandos de deploy
docker-push: ## Faz push da imagem para o registry
	@echo "📤 Enviando imagem para registry..."
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo "✅ Imagem enviada"

docker-tag: ## Cria uma tag para a imagem
	@read -p "Digite a versão (ex: v1.0.0): " version; \
	docker tag $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $(DOCKER_IMAGE_NAME):$$version
	@echo "✅ Tag criada"

# Info
info: ## Mostra informações do projeto
	@echo "📊 Informações do Projeto"
	@echo "------------------------"
	@echo "Imagem Docker: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"
	@echo "Binário: $(BINARY_NAME)"
	@echo ""
	@echo "Versão Go:"
	@go version
	@echo ""
	@echo "Dependências:"
	@go list -m all | head -5
