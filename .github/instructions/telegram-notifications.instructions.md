# Telegram Notifications Instruction
- Version: 202602142220
- Scope: GitHub Actions workflow for Telegram notifications in this repository.

## Files
- Workflow: `.github/workflows/telegram-notify.yml`
- Local token file (confidential): `.telegram.token`

## Mandatory Rules
- Never expose token values in source code, logs, docs, or commits.
- Never commit `.telegram.token`.
- Always use GitHub Secrets in CI/CD:
  - `TELEGRAM_BOT_TOKEN` (required)
  - `TELEGRAM_CHAT_ID` (required)
  - `TELEGRAM_THREAD_ID` (optional)
- Keep notification text plain for compatibility and readability.
- Keep URL as the final line of message block to enable Telegram preview.

## Event Coverage
- `push` (main, develop)
- `pull_request` (opened, reopened, synchronize, closed)
- `issues` (opened, reopened, edited, closed)
- `issue_comment` (created)
- `release` (published)
- `workflow_dispatch`

## Operational Notes
- If required secrets are missing, workflow must skip sending notification.
- Before changing message format or workflow behavior, read this instruction first.
- Manual validation can be executed via `workflow_dispatch`.
