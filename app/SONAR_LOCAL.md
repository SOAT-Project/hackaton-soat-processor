# 🔍 SonarQube Local Testing

Este diretório contém scripts para executar o SonarQube Scanner localmente, permitindo testar a análise de código e coverage antes de fazer push para o repositório.

## 📋 Pré-requisitos

- Docker instalado e rodando
- Token do SonarCloud configurado
- Go 1.25+ instalado

## 🔑 Configurar Token do SonarCloud

1. Acesse [SonarCloud](https://sonarcloud.io)
2. Clique no seu avatar → **My Account**
3. **Security** → **Generate Tokens**
4. Copie o token gerado
5. Exporte no terminal:

```bash
export SONAR_TOKEN="seu-token-aqui"
```

💡 **Dica:** Adicione ao seu `~/.zshrc` ou `~/.bashrc` para tornar permanente:

```bash
echo 'export SONAR_TOKEN="seu-token-aqui"' >> ~/.zshrc
source ~/.zshrc
```

## 🚀 Como Usar

### Opção 1: Usando Makefile (Recomendado)

```bash
cd app

# 1. Testar transformação de paths do coverage
make test-paths

# 2. Executar SonarQube Scanner localmente
make sonar-local
```

### Opção 2: Executando scripts diretamente

```bash
cd app

# 1. Testar transformação de paths
./test-coverage-paths.sh

# 2. Executar SonarQube Scanner
./run-sonar-local.sh
```

### Opção 3: Passo a passo manual

```bash
cd app

# 1. Gerar coverage
go test -coverprofile=coverage.out -covermode=atomic ./internal/...

# 2. Ver coverage original
head -n 10 coverage.out

# 3. Transformar paths (backup automático)
sed -i.bak 's|github.com/SOAT-Project/hackaton-soat-processor/||g' coverage.out

# 4. Verificar transformação
head -n 10 coverage.out

# 5. Executar scanner
docker run \
    --rm \
    -e SONAR_TOKEN="$SONAR_TOKEN" \
    -v "$(pwd):/usr/src" \
    sonarsource/sonar-scanner-cli:latest \
    -Dsonar.projectBaseDir=/usr/src \
    -Dsonar.verbose=true

# 6. Restaurar backup
mv coverage.out.bak coverage.out
```

## 📊 Scripts Disponíveis

### `test-coverage-paths.sh`

Testa a transformação de paths do arquivo de coverage para garantir que os paths correspondem aos arquivos reais do projeto.

**O que faz:**
- Gera o coverage se não existir
- Mostra os paths originais (com prefixo do módulo)
- Transforma os paths para paths relativos
- Verifica se cada path transformado corresponde a um arquivo real
- Mostra resumo de cobertura

**Uso:**
```bash
./test-coverage-paths.sh
```

**Output esperado:**
```
✅ internal/adapter/ffmpeg_processor.go
✅ internal/adapter/message_adapter.go
✅ internal/adapter/storage_adapter.go
✅ internal/application/domain/video_process.go
✅ internal/application/usecase/process_video_usecase.go
```

### `run-sonar-local.sh`

Executa o SonarQube Scanner localmente usando Docker, simulando o que acontece na pipeline do GitHub Actions.

**O que faz:**
- Verifica pré-requisitos (Docker, token, etc)
- Gera ou usa coverage existente
- Cria backup do coverage original
- Transforma os paths para formato esperado pelo SonarQube
- Executa o scanner via Docker
- Restaura o coverage original
- Mostra link para ver resultados no SonarCloud

**Uso:**
```bash
./run-sonar-local.sh
```

**Opções interativas:**
- Se coverage.out já existir, pergunta se deseja regerar

## 🔧 Solução de Problemas

### Erro: "Docker is not running"

**Causa:** Docker daemon não está ativo

**Solução:**
```bash
# Linux
sudo systemctl start docker

# macOS
open -a Docker

# Verificar
docker info
```

### Erro: "SONAR_TOKEN not set"

**Causa:** Variável de ambiente não configurada

**Solução:**
```bash
export SONAR_TOKEN="seu-token-aqui"
```

### Erro: "There are not enough lines to compute coverage"

**Causa:** Paths no coverage.out não correspondem aos arquivos do projeto

**Solução:** Este erro é exatamente o que os scripts corrigem! Execute:
```bash
./test-coverage-paths.sh
```

Se todos os arquivos mostrarem ✅, a transformação está correta.

### Coverage não aparece no SonarCloud

**Possíveis causas:**
1. Paths incorretos no coverage.out (use `test-coverage-paths.sh` para verificar)
2. Branch não configurado no SonarCloud
3. Token sem permissões adequadas
4. Arquivo coverage.out não foi enviado corretamente

**Debug:**
```bash
# 1. Verificar paths
./test-coverage-paths.sh

# 2. Verificar formato do coverage
head -20 coverage.out

# 3. Verificar se paths são relativos
grep "^internal/" coverage.out | head -5

# 4. Executar localmente com verbose
./run-sonar-local.sh
```

## 📈 Coverage Atual

Meta: **90%** do diretório `internal/`

Última medição:
- **Total:** 89.9%
- **Domain:** 100%
- **Storage Adapter:** 100%
- **Message Adapter:** 100%
- **Use Case:** 94.2%
- **FFmpeg Processor:** 78.2%

## 🔗 Links Úteis

- **SonarCloud Dashboard:** https://sonarcloud.io/project/overview?id=SOAT-Project_hackaton-soat-processor
- **SonarCloud Documentation:** https://docs.sonarcloud.io/
- **Go Coverage Tool:** https://go.dev/blog/cover
- **GitHub Actions Workflow:** `../.github/workflows/validation.yaml`

## 💡 Dicas

1. **Sempre teste localmente antes de fazer push:**
   ```bash
   make test-paths && make sonar-local
   ```

2. **Use o verbose mode para debug:**
   O script já inclui `-Dsonar.verbose=true`

3. **Verifique os logs do scanner:**
   Procure por mensagens como "Coverage report" e "Lines to cover"

4. **Compare com o resultado local:**
   ```bash
   go tool cover -func=coverage.out | tail -n 1
   ```

5. **Mantenha o coverage.out versionado apenas para debug:**
   O arquivo é gerado automaticamente na pipeline
