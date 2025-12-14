## GoProject (proxia) — Universal AI Gateway

A minimalist, stable gateway that unifies OpenAI- and Anthropic-style chat behind one endpoint. See `docs/specification.md` for the full spec.

## What it does
- Single endpoint: `POST /v1/chat` for all providers.
- Normalized responses: injects `provider`, fills `id`/`created`, OpenAI-like shape.
- Flexible auth: Bearer can be a gateway-issued key **or** a direct upstream key (OpenAI/Anthropic). BYOK via DB still works if configured.
- Dynamic routing: auto by model prefix, or force with `provider`; optional per-request `base_url` override.

## Quick start (backend)
Prereqs: Go 1.22+

```bash
cp example.env .env
# set upstream keys:
#   OPENAI_API_KEY=...
#   CLAUDE_API_KEY=...
go run ./backend
```

Smoke test:
```bash
curl -s http://localhost:8080/v1/health
curl -s http://localhost:8080/v1/config
```

### Calling /v1/chat
Auth options:
1) **Gateway key** (stored in DB or seeded by `GATEWAY_DEV_KEY` when no DB):
```
Authorization: Bearer <gateway_key>
```
2) **Direct upstream key** (no gateway lookup or rate limit):
```
Authorization: Bearer <your_openai_or_anthropic_key>
```

Body is OpenAI-style. Differences:
- Anthropic requires `max_tokens`.
- Optional `provider` to force routing; otherwise model prefix decides (`claude*` → Claude, `gpt*`/`openai`/`o1*` → OpenAI).
- Optional `base_url`/`provider_url` to override for this call; otherwise uses env defaults.
- Gateway-only fields (`provider`, `base_url`, `provider_url`) are stripped before sending upstream.

OpenAI example:
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

Claude example:
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer $ANTHROPIC_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":256}'
```

### Environment (backend)
- `OPENAI_API_KEY`, `CLAUDE_API_KEY`, `API_KEY` (fallback)
- `OPENAI_PROVIDER_URL` (default `https://api.openai.com/v1`)
- `ANTHROPIC_PROVIDER_URL` (default `https://api.anthropic.com`)
- `PROVIDER` (default `openai`)
- `PORT` (default `8080`), `LOG_LEVEL` (default `info`)
- `DATABASE_URL` (optional Postgres/Supabase for gateway keys)
- `GATEWAY_DEV_KEY` (optional dev key when no DB)
- `ENCRYPTION_KEY` (optional; needed only for DB-stored BYOK)

See `example.env` for a template.

### Provider status
- `openai`: implemented
- `claude` (Anthropic): implemented
- `gemini`: not implemented (501)

### Data layer
- With DB (`DATABASE_URL`): gateway keys in `public.api_keys` (hash = sha256(raw_key)); per-key rate limits honored.
- Without DB: in-memory store; set `GATEWAY_DEV_KEY` to seed one dev key.

### Web dashboard (MVP)
- Supabase login and key management live in `web/`.
- Run:
```bash
cd web
npm install
npm run dev
```
Open `http://localhost:3000`. Configure `web/.env.local` per `web/README.md`.

### Build & Docker
```bash
go build -o gateway ./backend
./gateway

docker build -t goproject .
docker run --rm -p 8080:8080 --env-file .env goproject
```

### Troubleshooting
- Use `go run .` from `backend/` to include all files.
- For Anthropic, always include `max_tokens`.
- Upstream errors bubble with provider HTTP status in `{"error":{"code":...,"message":"..."}}`.
- Increase verbosity with `LOG_LEVEL=debug`.

### License
MIT — see `LICENSE`.
