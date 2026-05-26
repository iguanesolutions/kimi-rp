# kimi-rp

Kimi Reverse Proxy is a lightweight HTTP reverse proxy for **Kimi K2.5 and K2.6** models that automatically adjusts sampling parameters (temperature, top_p) and thinking mode based on whether a thinking or non-thinking model is being used. It sits between your application and the backend LLM server (e.g., vLLM). It also provides `/tokenize` and `/v1/models` endpoints with full virtual model support.

## Core Functionality

This proxy's primary purpose is to:

1. **Accept requests for two virtual model names** (configured via `-thinking-model` and `-no-thinking-model`), rejecting all other model names with HTTP 400
2. **Set appropriate sampling parameters** automatically based on the model type (Kimi K2.5/K2.6 recommended values):
   - **Thinking mode**: `temperature=1.0`, `top_p=0.95`
   - **Instant (no-thinking) mode**: `temperature=0.6`, `top_p=0.95`
3. **Configure thinking mode** by setting `chat_template_kwargs.thinking`:
   - `thinking=true` for thinking model
   - `thinking=false` for no-thinking model
4. **Rewrite the model name** to the actual backend model name before forwarding to vLLM
5. **Fix vLLM response bugs** where non-thinking, non-streaming responses incorrectly place content in `reasoning_content` or `reasoning` fields instead of `content`
6. **Enrich `/v1/models` endpoint** by fetching backend models and exposing 2 virtual models with the same metadata
7. **Provide a `/tokenize` endpoint** that replaces virtual model names with the backend model name before forwarding to vLLM's `/tokenize`

## Installation

Requirements: Go 1.24.2 or later

```bash
go build -o kimi-rp .
```

## Usage

```bash
./kimi-rp \
  -target "http://127.0.0.1:8000" \
  -served-model "your-backend-model-name" \
  -thinking-model "kimi-k2.6-thinking" \
  -no-thinking-model "kimi-k2.6-instant"
```

Or using environment variables:

```bash
export KIMIRP_TARGET="http://127.0.0.1:8000"
export KIMIRP_SERVED_MODEL_NAME="your-backend-model-name"
export KIMIRP_THINKING_MODEL_NAME="kimi-k2.6-thinking"
export KIMIRP_NO_THINKING_MODEL_NAME="kimi-k2.6-instant"
./kimi-rp
```

## Configuration

Configure the proxy using command-line flags or environment variables:

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-listen` | `KIMIRP_LISTEN` | `0.0.0.0` | IP address to listen on |
| `-port` | `KIMIRP_PORT` | `9000` | Port to listen on |
| `-target` | `KIMIRP_TARGET` | `http://127.0.0.1:8000` | Backend target URL |
| `-loglevel` | `KIMIRP_LOGLEVEL` | `INFO` | Log level (COMPLETE, DEBUG, INFO, WARN, ERROR) |
| `-served-model` | `KIMIRP_SERVED_MODEL_NAME` | (required) | Backend model name to use in outgoing requests |
| `-thinking-model` | `KIMIRP_THINKING_MODEL_NAME` | (required) | Name of the thinking model (e.g., `kimi-k2.6-thinking`, `kimi-k2.5-thinking`) |
| `-no-thinking-model` | `KIMIRP_NO_THINKING_MODEL_NAME` | (required) | Name of the instant/no-thinking model (e.g., `kimi-k2.6-instant`, `kimi-k2.5-instant`) |
| `-enforce-sampling-params` | `KIMIRP_ENFORCE_SAMPLING_PARAMS` | `false` | Enforce sampling parameters, overriding client-provided values |

### Enforce Sampling Parameters

By default, the proxy only sets sampling parameters if they are not already present in the request. When `-enforce-sampling-params` is enabled, the proxy will **always override** client-provided sampling parameters with the predefined values for the detected mode.

## Request Routing

- **`GET /v1/models`**: Enriched (fetches backend models, validates served model, exposes 2 virtual models)
- **`POST /v1/chat/completions`**: Transformed (sampling params + thinking mode applied)
- **`POST /v1/completions`**: Model name validated and swapped (no sampling params or thinking mode — raw prompt completions bypass the chat template)
- **`POST /tokenize`**: Replaces virtual model names with backend model name and forwards to vLLM's `/tokenize`
- **All other paths**: Passed through unchanged to the backend

### vLLM Backend Requirements

For full functionality with thinking mode and tool calls using the Chat Completions API, the vLLM backend should be started with the following flags:

```bash
--reasoning-parser=qwen3                                  # Required for thinking/reasoning mode
--enable-auto-tool-choice --tool-call-parser=qwen3_coder  # Required for tool/function calls
```

## Tokenize API

The proxy provides a `/tokenize` endpoint that forwards tokenization requests to vLLM's `/tokenize`. The proxy replaces virtual model names with the backend served model name, then forwards the request body unchanged. Two modes:

- **`{"prompt": "..."}`** — raw text tokenization, forwarded as-is. No chat template is applied.
- **`{"messages": [...], "tools": [...]}`** — vLLM applies the model's chat template (`apply_chat_template`) then tokenizes the result. Messages and tools must be in Chat Completions API format.

## Health Check

- **`GET /health`**: Returns `{"status":"healthy"}` for Docker health checks

## Log Levels

The proxy supports the following log levels:

| Level | Description |
|-------|-------------|
| `COMPLETE` | Most verbose - includes full HTTP request/response dumps |
| `DEBUG` | Debug information including parameter application details |
| `INFO` | General operational information |
| `WARN` | Warning messages |
| `ERROR` | Error messages only |

When set to `COMPLETE`, the proxy will log full HTTP request and response bodies, which is useful for debugging but very verbose.

⚠️ **Privacy Warning**: LLM requests often contain sensitive or personal data (conversation history, personal information, confidential content). The `COMPLETE` log level will expose all this data in plaintext. Only enable it in secure, non-production environments or ensure logs are properly secured and retained temporarily.

## systemd Integration

The proxy includes native systemd support for production deployments:

- **Type**: `notify` - The proxy signals readiness to systemd automatically
- **Status Updates**: Sends periodic status updates to systemd showing processed request counts
- **Graceful Shutdown**: Properly signals systemd when stopping
- **Journald Logging**: Structured logging output is compatible with journald

Example systemd unit file:

```ini
[Unit]
Description=Kimi Reverse Proxy
After=network.target

[Service]
Type=notify
User=kimi-rp
Group=kimi-rp
ExecStart=/usr/local/bin/kimi-rp -served-model "your-backend-model" -thinking-model "kimi-k2.6-thinking" -no-thinking-model "kimi-k2.6-instant"
Restart=on-failure
Environment=KIMIRP_LOGLEVEL=INFO

[Install]
WantedBy=multi-user.target
```

⚠️ **Security Best Practice**: Always run the proxy under a dedicated, unprivileged user account (e.g., `kimi-rp`). Never run as root. Create the user with:
```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin kimi-rp
sudo chown kimi-rp:kimi-rp /usr/local/bin/kimi-rp
```

## Graceful Shutdown

The server supports graceful shutdown with a 3-minute timeout to allow in-flight requests to complete. Send `SIGINT` or `SIGTERM` to initiate shutdown. When running under systemd, the proxy will automatically signal the service manager when ready and during shutdown.

## License

MIT License - see [LICENSE](LICENSE) file for details.
