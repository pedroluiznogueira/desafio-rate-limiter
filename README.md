# Rate Limiter

Middleware HTTP em Go que limita requisições por IP ou por token (`API_KEY: <TOKEN>`), com
Redis para persistência via um `Strategy` (`internal/limiter.Strategy`) — trocar Redis por outro
mecanismo é só implementar essa interface e injetar em `cmd/server/main.go`, sem tocar no
middleware ou na lógica de negócio.

Um token configurado sempre sobrepõe o limite de IP (Regra de Ouro do desafio), mesmo que menor
que o limite de IP.

## Configuração (variáveis de ambiente / `.env`)

Veja `.env.example`:

- `RATE_LIMIT_IP` — máximo de requisições/segundo por IP (default 10).
- `RATE_LIMIT_TOKENS` — limites por token, formato `token:limite,token2:limite2`
  (ex: `abc123:100`).
- `RATE_LIMIT_BLOCK_DURATION` — tempo de bloqueio após exceder o limite (ex: `5m`).
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` — conexão com o Redis.

## Rodando

```
docker compose up --build
```

Sobe a API (porta 8080) e o Redis. Teste:

```
curl -i http://localhost:8080/
curl -i http://localhost:8080/ -H "API_KEY: abc123"
```

Excedendo o limite, a resposta é `429` com o corpo:
`you have reached the maximum number of requests or actions allowed within a certain time frame`

## Rodando os testes (via Docker, sem instalar Go localmente)

```
docker compose up --build test
```

Os testes sobem contra o Redis real do compose e demonstram: requisições dentro do limite
passam, exceder o limite bloqueia (e o bloqueio persiste além da janela de contagem de 1s), e um
token configurado com limite próprio continua funcionando mesmo com o IP já bloqueado
(precedência Token > IP).
