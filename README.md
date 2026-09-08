# Plug-and-Play Go Voice Calling Backend Module

Outbound voice calling for Go backends, powered by the official [Plivo Go SDK](https://github.com/plivo/plivo-go).

This repository is a standalone POC **and** a reusable LEGO-block module. Plivo-specific details stay behind a clean interface so another Go service (for example PhishSheriff) can integrate with minimal configuration and code changes.

> The module encapsulates Plivo-specific implementation details behind a clean interface so it can be integrated into another Go backend with minimal configuration and code changes.

---

## 1. Project Overview

| Mode | Purpose |
|------|---------|
| **Library** (`pkg/voice`) | Import into any Go backend and call `InitiateCall` |
| **POC binary** (`cmd/server`) | One-shot CLI call **or** tiny HTTP server for Docker/Kubernetes |

The consuming backend should not need to understand Plivo SDK internals, authentication, or request construction.

---

## 2. Architecture

```text
Application / Backend
        │
        ▼
Voice Calling Interface   (pkg/voice)
        │
        ▼
Plivo Implementation      (pkg/voice/plivo)
        │
        ▼
Plivo Voice API
```

Conceptual wiring:

```text
External Go Backend
│
▼
Voice Calling Module
│
▼
Plivo Go SDK
│
▼
Plivo Voice API
```

---

## 3. Plug-and-Play Integration

How another Go backend (e.g. PhishSheriff) integrates this module:

1. **Add the module** — clone this repo, or `go get` / replace with a private module path.
2. **Configure environment variables** — `PLIVO_AUTH_ID`, `PLIVO_AUTH_TOKEN`, `PLIVO_FROM_NUMBER`, `PLIVO_ANSWER_URL`.
3. **Initialize configuration** — load and validate env (reuse `internal/config` or map your own config into the Plivo client).
4. **Initialize the voice service** — `plivo.NewClient(...)`.
5. **Call** `InitiateCall(ctx, request)`.

### Integration example

```go
package example

import (
	"context"
	"log/slog"

	"github.com/sravan-chavan/Voice-Calling-POC/pkg/voice"
	plivovoice "github.com/sravan-chavan/Voice-Calling-POC/pkg/voice/plivo"
)

func NewVoiceService(authID, authToken, from, answerURL string, logger *slog.Logger) (voice.Service, error) {
	return plivovoice.NewClient(plivovoice.Config{
		AuthID:       authID,
		AuthToken:    authToken,
		FromNumber:   from,
		AnswerURL:    answerURL,
		AnswerMethod: "GET",
		Logger:       logger,
	})
}

func PlaceCall(svc voice.Service, to string) (*voice.InitiateCallResponse, error) {
	return svc.InitiateCall(context.Background(), voice.InitiateCallRequest{To: to})
}
```

Depend on `voice.Service` in business logic so a future provider can be swapped without rewriting callers.

---

## 4. Local Development

### Prerequisites

- Go **1.22+** (tested with Go 1.22+)
- A Plivo account with:
  - Auth ID / Auth Token
  - A purchased/verified caller ID (`PLIVO_FROM_NUMBER`)
  - An Answer URL that returns Plivo XML when the call is answered

### Install dependencies

```bash
go mod download
```

### Create `.env`

```bash
cp .env.example .env
# edit .env with your values
```

`.env` is gitignored and **optional in production** (use real environment variables instead).

### Run a one-shot outbound call

```bash
export $(grep -v '^#' .env | xargs)   # or: set -a; source .env; set +a
go run ./cmd/server -to=+14155552671
```

Or:

```bash
CALL_TO=+14155552671 go run ./cmd/server
```

### Run HTTP server mode (for local API testing)

```bash
go run ./cmd/server -serve -http-addr=:8080
```

```bash
curl -s localhost:8080/healthz
curl -s -X POST localhost:8080/v1/calls \
  -H 'Content-Type: application/json' \
  -d '{"to":"+14155552671"}'
```

### Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/server ./cmd/server
```

---

## 5. Docker

Build:

```bash
docker build -t voice-calling:latest .
```

Run (pass secrets via environment — never bake them into the image):

```bash
docker run --rm -p 8080:8080 \
  -e PLIVO_AUTH_ID \
  -e PLIVO_AUTH_TOKEN \
  -e PLIVO_FROM_NUMBER \
  -e PLIVO_ANSWER_URL \
  voice-calling:latest
```

One-shot call inside the container:

```bash
docker run --rm \
  -e PLIVO_AUTH_ID \
  -e PLIVO_AUTH_TOKEN \
  -e PLIVO_FROM_NUMBER \
  -e PLIVO_ANSWER_URL \
  voice-calling:latest \
  -to=+14155552671
```

### docker-compose

```bash
cp .env.example .env   # fill in values
docker compose up --build
```

---

## 6. Kubernetes

Manifests live under `deployments/kubernetes/`:

| File | Purpose |
|------|---------|
| `deployment.yaml` | Long-running HTTP server |
| `service.yaml` | ClusterIP Service |
| `configmap.yaml` | Non-sensitive config |
| `secret.example.yaml` | Secret template (no real credentials) |

### Configure secrets

```bash
kubectl create secret generic voice-calling-secrets \
  --from-literal=PLIVO_AUTH_ID='your-auth-id' \
  --from-literal=PLIVO_AUTH_TOKEN='your-auth-token'
```

### Deploy

```bash
# Edit configmap.yaml with your from-number and answer URL
kubectl apply -f deployments/kubernetes/configmap.yaml
kubectl apply -f deployments/kubernetes/deployment.yaml
kubectl apply -f deployments/kubernetes/service.yaml
```

Build/push the image to your registry and update `image:` in `deployment.yaml` before applying in a real cluster.

---

## 7. Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `PLIVO_AUTH_ID` | Yes | Plivo Auth ID |
| `PLIVO_AUTH_TOKEN` | Yes | Plivo Auth Token |
| `PLIVO_FROM_NUMBER` | Yes | Caller ID (Plivo number on your account), E.164 |
| `PLIVO_ANSWER_URL` | Yes | Absolute URL Plivo fetches when the call is answered |
| `PLIVO_ANSWER_METHOD` | No | `GET` (default) or `POST` |
| `CALL_TO` | No | Destination for one-shot CLI mode |
| `HTTP_ADDR` | No | If set (e.g. `:8080`), enables HTTP server mode |

Local development may use `.env`. Production (Docker / Kubernetes / CI) should inject environment variables directly — `.env` is not required.

---

## 8. API / Module Usage

### Public interface

```go
type Service interface {
    InitiateCall(ctx context.Context, req InitiateCallRequest) (*InitiateCallResponse, error)
}
```

### Request

```go
type InitiateCallRequest struct {
    To string // E.164, e.g. "+14155552671"
}
```

### Response

```go
type InitiateCallResponse struct {
    RequestUUID string
    Message     string
    APIID       string
    From        string
    To          string
}
```

### HTTP (POC server mode only)

| Method | Path | Body | Description |
|--------|------|------|-------------|
| `GET` | `/healthz` | — | Liveness/readiness |
| `POST` | `/v1/calls` | `{"to":"+E.164"}` | Initiate outbound call |

---

## 9. Call Flow

```text
Backend
→ Voice Module (validate + map request)
→ Plivo SDK (Calls.Create)
→ Plivo API
→ Destination Phone
→ On answer: Plivo fetches PLIVO_ANSWER_URL (XML instructions)
```

---

## 10. Security

- `.env` is listed in `.gitignore` and must never be committed
- Credentials are never hardcoded
- Logs redact auth tokens; auth IDs are masked in config dumps
- Kubernetes Secrets hold `PLIVO_AUTH_ID` / `PLIVO_AUTH_TOKEN`
- Docker images do not include `.env` or secrets

---

## 11. Troubleshooting

| Issue | What to check |
|-------|----------------|
| Invalid credentials | Verify `PLIVO_AUTH_ID` / `PLIVO_AUTH_TOKEN` from the Plivo console |
| Invalid phone format | Use E.164 (`+` and country code), e.g. `+14155552671` |
| Caller ID not authorized | `PLIVO_FROM_NUMBER` must be a Plivo number on the same account |
| Answer URL problems | URL must be publicly reachable and return valid Plivo XML |
| Network failures | Outbound HTTPS to Plivo must be allowed from the runtime |
| Config missing at startup | Application fails fast listing missing env vars |

---

## Project layout

```text
Voice-Calling-POC/
├── cmd/server/                 # POC CLI + optional HTTP server
├── internal/config/            # Env loading + validation
├── pkg/voice/                  # Public interface, types, validation
│   └── plivo/                  # Plivo SDK adapter
├── deployments/kubernetes/     # K8s manifests
├── tests/                      # Interface mockability examples
├── examples/reference/         # Preserved original BulkCalls.go sample
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

### Note on prior artifacts

- `examples/reference/BulkCalls.go` — original reference snippet (placeholders; malformed `To` field). Kept for history; **do not run as-is**.
- `~/plivo-voice-calling/test-call.sh` — separate Agent Flow webhook test script (outside this repo). This module uses the **Voice Call API** via the Go SDK, not that webhook path.

---

## License

Internal POC — use according to your organization's policies.
