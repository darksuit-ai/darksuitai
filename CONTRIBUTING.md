# Contributing to DarkSuitAI

Thanks for your interest — contributions of all kinds are welcome: features, bug
fixes, docs, examples, and new provider integrations.

## Getting started

```bash
git clone https://github.com/darksuit-ai/darksuitai
cd darksuitai
go mod tidy
go build ./...
go test ./...
```

Requires **Go 1.24+**. To run the live example, copy `.env.example` to `.env`,
add a provider key, and run `go run ./doc`.

## Development workflow

1. Fork and create a branch: `git checkout -b feat/my-change`.
2. Make your change with tests where practical.
3. Before pushing:
   ```bash
   gofmt -l .        # must print nothing
   go vet ./...
   go test ./...
   ```
4. Open a PR against `main` and fill in the template.

## Guidelines

- **Keep it lean.** DarkSuitAI's value is a small, sharp surface. Prefer clear
  code over clever abstraction.
- **Preserve the public API** in `darksuitai.go` unless a change is intentional
  and documented. Provider internals live in `internal/llms/<provider>` and
  should keep the shared `llms.LLM` contract.
- **No secrets in code or tests.** Read keys from the environment.
- **Format with `gofmt`** and pass `go vet`.
- **Document exported symbols** with Go doc comments.

## Good places to help

- New provider integrations (mirror `internal/llms/openaicompat` or the
  Anthropic package).
- Native tool calling for the OpenAI-compatible providers.
- Token-usage telemetry (surface SDK usage into the `Observer`).
- More examples under `doc/`.

See the [good first issues](https://github.com/darksuit-ai/darksuitai/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

## Reporting bugs / requesting features

Use the issue templates. For bugs, include your Go version, provider, a minimal
repro, and what you expected vs. what happened.

## Code of conduct

Be respectful and constructive. We want this to be a welcoming project.
