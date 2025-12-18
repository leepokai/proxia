## GoProject (proxia) — Universal AI Gateway

A minimalist, stable gateway that unifies OpenAI- and Anthropic-style chat behind one endpoint. See `docs/specification.md` for the full spec.

## Recent Updates

This project has undergone significant improvements:

- **Test Infrastructure**: Comprehensive test suite added in `backend/utils/tests/` covering crypto utilities, response normalization, and error handling
- **Code Organization**: Tests moved to dedicated `tests/` subdirectory for better separation of concerns
- **Architecture Refinement**: Improved middleware, rate limiting, and store abstractions
- **Web Dashboard**: Full Next.js frontend with Supabase authentication and key management

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

#### Authentication Options

**1. Gateway Key (Recommended)**

Use a single gateway key created via the web dashboard to access multiple providers (OpenAI, Claude, etc.). This is the recommended approach for unified access control and rate limiting.

```
Authorization: Bearer <gateway_key>
```

**Key Benefits:**
- ✅ Single key for all providers (Claude, GPT, etc.)
- ✅ Automatic routing based on model name
- ✅ Rate limiting per key
- ✅ Centralized key management via web dashboard
- ✅ Optional BYOK (Bring Your Own Key) support for encrypted provider credentials

**Example: Using one gateway key for different models**

```bash
# Use the same gateway key for OpenAI
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <your_gateway_key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'

# Use the same gateway key for Claude
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <your_gateway_key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":256}'
```

The gateway automatically routes to the correct provider based on the model name:
- `gpt-*`, `openai`, `o1-*` → OpenAI
- `claude-*` → Anthropic Claude

You can also explicitly specify the provider:
```json
{"provider": "openai", "model": "gpt-4o-mini", ...}
```

**2. Direct Upstream Key (Pass-through)**

Use your own OpenAI or Anthropic API key directly (no gateway lookup or rate limiting):

```
Authorization: Bearer <your_openai_or_anthropic_key>
```

#### Request Body

Body follows OpenAI-style format. Key differences:
- **Anthropic requires `max_tokens`** (must be included for Claude models)
- **Optional `provider`**: Force routing to a specific provider (e.g., `"provider": "openai"`)
- **Optional `base_url`/`provider_url`**: Override the provider URL for this request
- Gateway-only fields (`provider`, `base_url`, `provider_url`) are automatically stripped before forwarding to upstream providers

#### Examples

**OpenAI with gateway key:**
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <gateway_key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

**Claude with gateway key:**
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer <gateway_key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"hi"}],"max_tokens":256}'
```

**Direct upstream key (OpenAI):**
```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

**Direct upstream key (Claude):**
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

## Project Architecture

### Backend (Go)

The backend is organized into several key components:

#### Core Components
- **`main.go`**: Application entry point, initializes configuration, providers, store, and rate limiter
- **`router.go`**: HTTP routing and request handling logic
- **`config.go`**: Configuration management and environment variable loading

#### Handlers (`handlers/`)
- **`provider.go`**: AI provider interface definition
- **`openai.go`**: OpenAI API handler implementation
- **`claude.go`**: Anthropic Claude API handler implementation
- **`gemini.go`**: Gemini handler stub (planned for future implementation)

#### Authentication & Authorization (`auth/`)
- **`middleware.go`**: Bearer token authentication middleware
  - Supports gateway-issued keys (with rate limiting)
  - Supports direct upstream provider keys (pass-through mode)
  - Context-based key injection for downstream handlers

#### Rate Limiting (`ratelimit/`)
- **`limiter.go`**: Per-key rate limiting using `golang.org/x/time/rate`
  - Keyed limiter with configurable RPS and burst
  - Currently basic implementation; enhanced features planned

#### Data Layer (`store/`)
- **`models.go`**: Data models and Store interface
- **`memory.go`**: In-memory store implementation (for development)
- **`pg.go`**: PostgreSQL/Supabase store implementation (production)
  - API key lookup with SHA-256 hashing
  - Provider credential storage with encryption support
  - Per-key rate limit configuration

#### Utilities (`utils/`)
- **`crypto.go`**: Encryption/decryption utilities for BYOK (AES-256-GCM)
- **`logger.go`**: Structured logging with configurable levels
- **`response.go`**: HTTP response helpers and response normalization
- **`errors.go`**: Standardized error response formatting
- **`tests/`**: Comprehensive test suite (separate package for better isolation)
  - `crypto_test.go`: Tests for encryption/decryption and key derivation
  - `response_test.go`: Tests for response normalization and HTTP writers

### Frontend (Next.js)

Located in `web/`:
- **Next.js 14+** with App Router
- **Supabase** integration for authentication
- **Key Management UI**: Create and manage gateway API keys
- **Provider Credentials**: Store encrypted upstream provider keys (BYOK)
- **Playground**: Interactive chat interface for testing

### Database (Supabase/PostgreSQL)

- **Migrations**: Located in `supabase/migrations/`
- **Schema**: User authentication, API keys, provider credentials
- **Encryption**: Optional encryption for stored provider credentials

### Request Flow

1. **Authentication**: Middleware validates Bearer token
   - If gateway key → lookup in store, apply rate limiting
   - If upstream key → pass through directly (no rate limiting)
   
2. **Routing**: Determine provider from `provider` hint or model prefix
   - Model-based: `gpt-*`, `openai`, `o1-*` → OpenAI; `claude-*` → Anthropic
   - Explicit: `{"provider": "openai"}` or `{"provider": "claude"}`
   - **Key feature**: Same gateway key works for all providers!
   
3. **Key Selection**: Priority order:
   - Direct upstream key (if provided as Bearer token)
   - BYOK from database (if gateway key + encryption enabled)
   - Server-level configured keys (from environment variables)
   
4. **Request Forwarding**: Strip gateway-only fields (`provider`, `base_url`, `provider_url`), forward to provider
   
5. **Response Normalization**: Inject `provider`, fill `id`/`created`, ensure consistent OpenAI-like shape

**Example Flow with Gateway Key:**
```
User → POST /v1/chat with gateway_key + model="gpt-4o-mini"
  → Gateway validates key, applies rate limit
  → Routes to OpenAI handler (based on model prefix)
  → Uses server-level OPENAI_API_KEY or BYOK
  → Forwards request to OpenAI API
  → Normalizes response with provider="openai"
  → Returns unified response
```

### Provider status
- `openai`: ✅ implemented
- `claude` (Anthropic): ✅ implemented
- `gemini`: 🚧 planned (not yet implemented)

**Note**: Gemini support and enhanced rate limiting features are planned for future releases.

### Data layer
- **With DB** (`DATABASE_URL`): Gateway keys stored in `public.api_keys` (hash = sha256(raw_key)); per-key rate limits honored.
- **Without DB**: In-memory store; set `GATEWAY_DEV_KEY` to seed one dev key for development.

### Testing

Run tests from the backend directory:
```bash
cd backend
go test ./...
```

Tests are located in `backend/utils/tests/` as a separate package to ensure proper isolation and test external API boundaries.

### Web dashboard (MVP)

The web dashboard provides a user-friendly interface for managing your gateway:

- **Authentication**: Supabase-based user login
- **API Key Management**: Create and manage gateway API keys
  - Generate keys that work with all providers (OpenAI, Claude, etc.)
  - Set per-key rate limits
  - View key usage and status
- **Provider Credentials (BYOK)**: Store encrypted upstream provider keys
- **Playground**: Interactive chat interface for testing different models

**Getting Started:**
```bash
cd web
npm install
npm run dev
```
Open `http://localhost:3000`. Configure `web/.env.local` per `web/README.md`.

**Create Your First Gateway Key:**
1. Log in via the web dashboard
2. Navigate to the Keys page
3. Create a new API key
4. Use this single key to access all providers (OpenAI, Claude, etc.)

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
