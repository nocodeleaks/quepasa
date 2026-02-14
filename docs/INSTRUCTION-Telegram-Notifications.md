# Telegram Notifications Instruction (GitHub -> Telegram)

This instruction document defines how Telegram notifications are configured and maintained for GitHub events.

## Scope

Workflow file: `.github/workflows/telegram-notify.yml`

Notification events:
- `push` (main, develop)
- `pull_request` (opened, reopened, synchronize, closed)
- `issues` (opened, reopened, edited, closed)
- `issue_comment` (created)
- `release` (published)
- `workflow_dispatch`

## Required GitHub Secrets

Repository path:
`Settings` -> `Secrets and variables` -> `Actions` -> `New repository secret`

Required:
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`

Optional:
- `TELEGRAM_THREAD_ID` (Telegram Topics)

## Local Token Reference (Confidential)

For local operations in this workspace, the Telegram bot token is stored in:

- `.telegram.token`

Rules:
- Never copy the token value into source code.
- Never commit the token file.
- Keep this file local-only on the developer machine.
- Use GitHub Secrets for CI/CD instead of local token files.

## Maintenance Notes

- If required secrets are missing, the workflow skips notification.
- Message format is plain text for compatibility and readability.
- URLs are sent at the end of message blocks for automatic Telegram preview.
- Manual validation can be triggered with `workflow_dispatch`.
