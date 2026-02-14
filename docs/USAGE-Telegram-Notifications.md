# Telegram Notifications (GitHub -> Telegram)

This project includes an automated GitHub Actions workflow to notify a Telegram group whenever important GitHub events happen.

## What it sends

Workflow file: `.github/workflows/telegram-notify.yml`

It sends notifications for:
- `push` (main, develop)
- `pull_request` (opened, reopened, synchronize, closed)
- `issues` (opened, reopened, edited, closed)
- `issue_comment` (created)
- `release` (published)
- `workflow_dispatch`

## Required GitHub Secrets

In your repository settings:

`Settings` -> `Secrets and variables` -> `Actions` -> `New repository secret`

Create:
- `TELEGRAM_BOT_TOKEN` (required)
- `TELEGRAM_CHAT_ID` (required)
- `TELEGRAM_THREAD_ID` (optional, for Telegram topics)

## How to get values

### 1) TELEGRAM_BOT_TOKEN
- Create a bot with [@BotFather](https://t.me/BotFather)
- Copy the bot token

### 2) TELEGRAM_CHAT_ID
- Add your bot to the target group
- Send at least one message in that group
- Open:
  - `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
- Find `chat.id` from the group message payload (usually negative for groups)

### 3) TELEGRAM_THREAD_ID (optional)
- Only if your group uses Topics
- Read `message_thread_id` in Telegram updates for the target topic

## Notes

- If required secrets are missing, the workflow job is skipped.
- Notifications are plain text by default for maximum compatibility.
- You can trigger a manual test from the Actions tab using `workflow_dispatch`.
