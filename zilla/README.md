# Zilla

Runs [Zilla](https://github.com/aklivity/zilla) as an OpenAI-compatible gateway in front of the [mocker](../mocker/README.md), so `benchmark.go` can measure it the same way it measures Bifrost: same payloads, same load, same `results.json`.

| File | What it is |
| --- | --- |
| `zilla.yaml` | Gateway config: `POST /v1/chat/completions` on port 7161, `gpt-*` models routed to the OpenAI upstream at `MOCKER_URL` |
| `docker-compose.yml` | Runs `ghcr.io/aklivity/zilla` (version from `ZILLA_VERSION`, default `2.6.0`) with that config |

This is the Zilla equivalent of a Bifrost setup with a single OpenAI provider whose base URL is the mocker: one upstream, no retries, no caching, no logging, and no API key required from the caller (`benchmark.go` sends no `Authorization` header to gateways).

## Run

**1. Start the mocker** on the host:

```bash
cd mocker && go run main.go -port 8000
```

**2. Start Zilla:**

```bash
cd zilla && docker compose up -d --wait
```

Zilla reaches the mocker at `http://host.docker.internal:8000` by default. Set `MOCKER_URL` to point it elsewhere, e.g. `MOCKER_URL=http://10.0.0.12:8000 docker compose up -d --wait`.

**3. Add the port to `.env`** in the repo root:

```env
ZILLA_PORT=7161
```

**4. Run the benchmark:**

```bash
./benchmark -provider zilla -rate 500 -duration 30
```

To compare against Bifrost, run both with the same flags on the same machine, one at a time:

```bash
./benchmark -provider bifrost -rate 5000 -duration 60 -output results.json
./benchmark -provider zilla   -rate 5000 -duration 60 -output results.json
```

Both land in `results.json`, keyed `bifrost` and `zilla`.

## Check it by hand

```bash
curl -X POST http://localhost:7161/v1/chat/completions -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"Hi"}],"model":"gpt-4o-mini"}'
```

## Notes

- **Message field order.** Zilla rejects a chat message whose `content` comes before `role` (`400`, "The 'role' parameter must precede 'content'"). `benchmark.go` builds messages from a struct so `role` always comes first; every provider receives the same bytes.
- **Server memory.** `benchmark.go` samples the process listening on `ZILLA_PORT`. With Docker's port mapping that is Docker's proxy (on macOS, the Docker backend), not Zilla, and the same is true for Bifrost in Docker. For real numbers on Linux, run the gateway with `network_mode: host` and the benchmark as root, so the port belongs to the gateway's own process.
- **Workers.** Zilla starts one worker per CPU by default. Add `"-w", "<n>"` to `command` in `docker-compose.yml` to pin it.
