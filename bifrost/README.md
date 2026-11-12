# Bifrost

The Bifrost setup used for the Bifrost vs [Zilla](../zilla/README.md) comparison: one OpenAI provider whose base URL is the [mocker](../mocker/README.md).

| File | What it is |
| --- | --- |
| `config.json` | One OpenAI provider pointed at the mocker on `host.docker.internal:8000`, no retries, logging and stores off, no auth on inference |
| `docker-compose.yml` | Runs `maximhq/bifrost` (version from `BIFROST_VERSION`, default `v2.3.2`) with that config on port 8080 |

## Settings

| Setting | Value | Source |
| --- | --- | --- |
| `client.initial_pool_size` | 10000 | [Bifrost benchmarks](https://docs.getbifrost.ai/benchmarking/getting-started), t3.medium |
| `concurrency_and_buffer_size.buffer_size` | 15000 | same |
| `concurrency_and_buffer_size.concurrency` | 1000 | Bifrost's default (`DefaultConcurrency`); the benchmark pages don't give one |
| `network_config.max_retries` | 0 | no retries, like Zilla |

For the t3.xlarge numbers, Bifrost uses `initial_pool_size` 15000 and `buffer_size` 20000.

## Run

**1. Start the mocker** on the host:

```bash
cd mocker && go run main.go -port 8000
```

**2. Start Bifrost:**

```bash
cd bifrost && docker compose up -d --wait
```

To reach the mocker somewhere else, change `network_config.base_url` in `config.json`.

**3. Run the benchmark** with `BIFROST_PORT=8080` in `.env`:

```bash
./benchmark -provider bifrost -rate 500 -duration 30
```
