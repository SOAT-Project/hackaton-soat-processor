# Observabilidade

Este documento descreve a stack de observabilidade do Video Processor Worker.

## Componentes

### 1. Logs Estruturados (Zap)
- Logs em formato JSON para produção
- Logs coloridos para desenvolvimento
- Níveis: DEBUG, INFO, WARN, ERROR
- Campos contextuais automáticos

### 2. Métricas (Prometheus)
Endpoint: `http://localhost:8080/metrics`

#### Métricas Disponíveis

**Contadores:**
- `worker_messages_processed_total{status}` - Total de mensagens processadas
- `worker_videos_processed_total{status}` - Total de vídeos processados
- `worker_errors_total{type}` - Total de erros por tipo
- `worker_s3_operations_total{operation,status}` - Operações S3
- `worker_sqs_operations_total{operation,status}` - Operações SQS

**Histogramas:**
- `worker_processing_duration_seconds{status}` - Duração do processamento
- `worker_file_size_bytes{type}` - Tamanho dos arquivos

**Gauges:**
- `worker_frames_extracted_last` - Frames extraídos do último vídeo
- `worker_messages_active` - Mensagens em processamento

### 3. Health Checks
- `http://localhost:8080/health` - Status de saúde
- `http://localhost:8080/ready` - Status de prontidão

### 4. Prometheus
- UI: `http://localhost:9090`
- Scrape interval: 10s
- Retenção: 15 dias

### 5. Grafana
- UI: `http://localhost:3000`
- Usuário padrão: `admin`
- Senha padrão: `admin123`

## Dashboard

O dashboard "Video Processor Worker Overview" inclui:

1. **Processing Rate** - Taxa de processamento (vídeos/min)
2. **Success Rate** - Porcentagem de sucesso
3. **Processing Duration** - Percentis p50, p95, p99
4. **Frames Extracted** - Frames do último vídeo
5. **Errors by Type** - Erros categorizados
6. **S3 Operations** - Operações no S3
7. **Messages Processed** - Mensagens processadas

## Iniciar

```bash
# Subir toda a stack
docker compose up -d

# Ver logs do worker
docker compose logs -f worker

# Ver logs do Prometheus
docker compose logs -f prometheus

# Ver logs do Grafana
docker compose logs -f grafana
```

## Acessar Serviços

- **Worker Metrics**: http://localhost:8080/metrics
- **Worker Health**: http://localhost:8080/health
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

## Configuração

### Variáveis de Ambiente

```bash
# .env
ENVIRONMENT=production  # ou development
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin123
```

### Desenvolvimento Local

Para ambiente de desenvolvimento, configure:

```bash
ENVIRONMENT=development
```

Isso ativa:
- Logs coloridos no console
- Nível de log DEBUG
- Timestamps mais legíveis

## Queries Úteis no Prometheus

```promql
# Taxa de sucesso
rate(worker_videos_processed_total{status="success"}[5m]) / 
(rate(worker_videos_processed_total{status="success"}[5m]) + 
 rate(worker_videos_processed_total{status="error"}[5m])) * 100

# Duração média
rate(worker_processing_duration_seconds_sum[5m]) / 
rate(worker_processing_duration_seconds_count[5m])

# Erros por minuto
rate(worker_errors_total[1m]) * 60

# Mensagens ativas
worker_messages_active
```

## Alertas (Futuro)

Exemplos de alertas que podem ser configurados:

- Taxa de erro > 10%
- Duração de processamento > 5min (p95)
- Worker sem processar mensagens por 5min
- Erros de S3/SQS aumentando

## Troubleshooting

### Prometheus não está coletando métricas

1. Verifique se o worker está rodando: `docker compose ps`
2. Verifique os logs: `docker compose logs worker`
3. Teste o endpoint: `curl http://localhost:8080/metrics`
4. Verifique a configuração: `cat observability/prometheus.yml`

### Grafana não mostra dados

1. Verifique se o Prometheus está funcionando: http://localhost:9090
2. Verifique o datasource no Grafana: Configuration > Data Sources
3. Execute uma query de teste no Prometheus
4. Recarregue o dashboard

### Logs não aparecem estruturados

Verifique a variável `ENVIRONMENT`. Se for `development`, os logs serão coloridos.
Para JSON, use `ENVIRONMENT=production`.

## Performance

A stack de observabilidade adiciona:
- ~10-20MB de memória ao worker
- <1% de CPU overhead
- ~100KB/min de dados de métricas

## Backup

Dados persistentes:
- Prometheus: volume `prometheus-data`
- Grafana: volume `grafana-data`

Para backup:

```bash
docker run --rm -v prometheus-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/prometheus-backup.tar.gz /data

docker run --rm -v grafana-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/grafana-backup.tar.gz /data
```

## Próximos Passos

- [ ] Configurar alertas no Prometheus
- [ ] Integrar com Alertmanager
- [ ] Adicionar tracing com Jaeger
- [ ] Exportar logs para Loki
- [ ] Dashboard de custos AWS
