## Learned User Preferences

- When the user says "yolo" or "build and test until done without asking", skip per-step review and finish all items in one pass, building and testing along the way.
- Commit messages in these repos are conventional commit subjects only (for example `feat(gmail): ...`, `fix(expenses): ...`), with no body.

## Learned Workspace Facts

- `go` is not on the agent shell PATH (also not in a login shell). Use the local SDK: `export PATH=$HOME/sdk/go1.25.9/bin:$PATH` before `go build` or `go test` in housematee-tgbot and expensemate-tgbot.
- `.local/` is git-ignored in both repos (global ignore plus repo `.gitignore`). It holds local secrets and sample data such as OAuth client files, tokens, and sample bank emails. Never commit or print its contents.
- In expensemate-tgbot, the live entry point is `cmd/bot/main.go` with code under `internal/bot`, `internal/handler`, `internal/service`, and `internal/repository`. `internal/bots/`, `internal/servers/`, and `pkg/` are legacy code that `main.go` does not import.
- In expensemate-tgbot, `configs/prod.yaml` and `configs/local.yaml` are git-ignored and hold real credentials (including the `gmail` block in `prod.yaml`). Never print their values.
- expensemate-tgbot runs in production on Fly.io as one always-on machine with no HTTP server (`fly.toml`). Run only one bot instance at a time: a second instance causes Telegram getUpdates conflicts and can ingest the same Gmail email twice, so keep the `gmail` block out of `configs/local.yaml` while prod runs.
- The Gmail OAuth app for expensemate-tgbot must stay "In production", not "Testing". In Testing, refresh tokens expire after 7 days.
- To ship new code for expensemate-tgbot, run `make deploy` (`flyctl deploy`, builds from the local working tree). `make redeploy` only restarts the current release (`flyctl apps restart`) with no build. When the user says "redeploy" after code changes, they mean ship the new code, so use `make deploy`. Check startup with `flyctl logs --no-tail`. The Fly "not listening" warning is expected because the bot has no HTTP server.
