# Configuração do SonarQube - Diretório Internal

## Resumo das Alterações

Este documento descreve as configurações realizadas para que o SonarQube analise apenas o diretório `internal`.

## Arquivos Configurados

### 1. sonar-project.properties
- **Localização**: `/app/sonar-project.properties`
- **Configuração**: 
  - `sonar.sources=internal` - Apenas código do diretório internal
  - `sonar.tests=internal` - Apenas testes do diretório internal
  - `sonar.go.coverage.reportPaths=coverage.out` - Relatório de cobertura

### 2. Makefile
- **Localização**: `/app/Makefile`
- **Novos Targets**:
  - `make test-internal` - Executa testes apenas do diretório internal
  - `make test-coverage-internal` - Gera coverage apenas do internal
  - `make sonar` - Prepara análise do SonarQube (roda test-coverage-internal)

### 3. GitHub Actions Workflow
- **Localização**: `/.github/workflows/validation.yaml`
- **Configuração**: Já estava configurado para rodar testes apenas do internal
  - `go test -coverprofile=coverage.out -covermode=atomic ./internal/...`

## Como Usar

### Localmente

```bash
# Executar testes do internal com coverage
cd app
make test-coverage-internal

# Preparar para análise SonarQube
make sonar

# Executar SonarScanner (se instalado)
sonar-scanner
```

### No CI/CD

O workflow do GitHub Actions já está configurado para:
1. Executar testes apenas do diretório internal
2. Gerar o arquivo coverage.out
3. Enviar para o SonarCloud

## Cobertura Atual

- **Total (internal)**: 25.1%
- **usecase**: 36.0%
- **adapter**: 0.0% (sem testes)
- **domain**: 0.0% (sem testes)
- **port**: sem arquivos de teste

## Próximos Passos para Melhorar Cobertura

1. Criar testes para os adapters (storage, message, ffmpeg)
2. Criar testes para o domain (ToSuccessMessage, ToErrorMessage)
3. Melhorar cobertura do usecase (atualmente em 36%)

## Estrutura de Diretórios Analisada

```
app/internal/
├── adapter/
│   ├── ffmpeg_processor.go (0% coverage)
│   ├── message_adapter.go (0% coverage)
│   └── storage_adapter.go (0% coverage)
├── application/
│   ├── domain/
│   │   └── video_process.go (0% coverage)
│   └── usecase/
│       ├── process_video_usecase.go (36% coverage)
│       └── process_video_usecase_test.go
└── port/
    ├── message_port.go
    ├── storage_port.go
    └── video_processor_port.go
```

## Benefícios

1. ✅ SonarQube analisa apenas código de negócio (internal)
2. ✅ Não polui análise com código de infraestrutura (cmd, pkg)
3. ✅ Foco na qualidade do código core da aplicação
4. ✅ Métricas mais precisas e relevantes
5. ✅ Facilita identificação de pontos que precisam de testes
