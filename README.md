### GoProject (proxia) — Universal AI Gateway

A minimalist, stable, and fast gateway that unifies access to multiple AI providers (OpenAI, Gemini, Claude, Azure OpenAI, …) behind a single, provider-agnostic API.

Refer to the full spec in `docs/specification.md`.

### Features
- **Single endpoint**: Unified `POST /v1/chat` across providers
- **Provider-agnostic**: Switch via `.env` without code changes
- **Simple & stable**: Minimal moving parts, clear error contracts
- **Ready to deploy**: Dockerfile included

### Quick start
Prerequisites: Go 1.22+

1) Configure environment
```bash
cp example.env .env
# Edit .env to set API_KEY (and optionally PROVIDER, PROVIDER_URL, PORT, LOG_LEVEL)
# Note: If PROVIDER=openai and PROVIDER_URL is empty, it defaults to https://api.openai.com/v1
```

2) Run the server
```bash
go run .
```

3) Smoke test
```bash
curl -s http://localhost:8080/v1/health
curl -s http://localhost:8080/v1/config
curl -s -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hello from GoProject"}],"temperature":0.2}'
```

If `API_KEY` is missing, `POST /v1/chat` returns:
```json
{"error":{"code":401,"message":"missing API key"}}
```

If you select an unimplemented provider (e.g. `gemini`, `claude`), the gateway returns:
```json
{"error":{"code":501,"message":"gemini provider not implemented"}}
```

### Configuration (.env)
- `PROVIDER`: `openai` (default) | `gemini` | `claude`
- `PROVIDER_URL`: Base URL of provider.
  - If it ends with `/v1`, gateway calls `/chat/completions`
  - Otherwise gateway calls `/v1/chat/completions`
- `API_KEY`: Provider API key
- `PORT`: HTTP port (default `8080`)
- `LOG_LEVEL`: `debug` | `info` | `warn` | `error` (default `info`)

See `example.env` for a template.

### API
- `GET /v1/health` → `{ "status": "ok", "uptime": "123s" }`
- `GET /v1/config` → Current provider, URL, port, log level
- `POST /v1/chat` → Forwards to configured provider and normalizes response:
  - Adds `"provider"`, ensures `"id"`, `"created"`, and default `"object":"chat.completion"`
  - On upstream errors, returns a uniform error with the upstream HTTP status and message when available

Example request:
```json
{
  "model": "gpt-4o-mini",
  "messages": [
    {"role": "user", "content": "Say hello from GoProject"}
  ],
  "temperature": 0.2
}
```

### Build
```bash
go build -o gateway .
./gateway
```

### Docker
```bash
docker build -t goproject .
docker run --rm -p 8080:8080 --env-file .env goproject
```

### Provider status
- `openai`: implemented
- `gemini`: returns HTTP 501 Not Implemented
- `claude`: returns HTTP 501 Not Implemented

### Project structure
```
.
├── main.go
├── router.go
├── config.go
├── handlers/
│   ├── provider.go
│   ├── openai.go
│   ├── gemini.go
│   └── claude.go
├── utils/
│   ├── logger.go
│   ├── errors.go
│   └── response.go
├── docs/
│   └── specification.md
├── example.env
├── Dockerfile
├── LICENSE
└── go.mod
```

### Troubleshooting
- Use `go run .` (not `go run main.go`) to compile all files in the module.
- If you see provider errors, verify `.env` contains valid `API_KEY` and `PROVIDER_URL`.
- Increase log detail with `LOG_LEVEL=debug`.
- Upstream provider errors will be surfaced with the provider's HTTP status (e.g. 400/401/429) and a concise message in `{ "error": { "code": ..., "message": "..." } }`.

### License
MIT — see `LICENSE`.
