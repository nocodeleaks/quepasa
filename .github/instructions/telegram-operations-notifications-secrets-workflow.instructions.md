# Telegram Notifications Instruction

## Scope
- Target workflow: `.github/workflows/telegram-notify.yml`
- Local confidential file: `.telegram.token`
- Telegram group chat ID for notifications: `-1001725217387`

## Agent Handling Rules
- When user requests any Telegram-specific operation, treat this file as the source of truth.
- If bot handling is required, use local credential source `.telegram.token` and configured chat ID.
- Keep this instruction file updated whenever Telegram bot credentials or posting target changes.

## Security Rules
- Never expose token values in code, logs, docs, commits, or pull requests.
- Never commit `.telegram.token`.
- Use GitHub Secrets only:
  - `TELEGRAM_BOT_TOKEN` (required)
  - `TELEGRAM_CHAT_ID` (required)
  - `TELEGRAM_THREAD_ID` (optional)

## Notification Rules
- Message text must be plain text.
- URL must be the final line of each message block.
- If required secrets are missing, skip send.

## Event Coverage
- `push` (main, develop)
- `pull_request` (opened, reopened, synchronize, closed)
- `issues` (opened, reopened, edited, closed)
- `issue_comment` (created)
- `release` (published)
- `workflow_dispatch`

## Validation
- Trigger manual run via `workflow_dispatch` after format changes.
- Confirm workflow success and Telegram delivery.
