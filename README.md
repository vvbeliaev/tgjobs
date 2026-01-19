# JobHunter 🕵️‍♂️

Automated job search via Telegram. This project listens to your Telegram channels and chats, filters messages using keywords, extracts structured vacancy data via LLM, and helps generate personalized "first touch" messages.

## How It Works

1.  **Collector**: A `gotd`-based userbot that listens to all incoming messages in your Telegram account.
2.  **Filtering**: Messages pass through a fast keyword filter (Go).
3.  **Job Module**:
    *   **Extraction**: LLM (GPT) analyzes the text to determine if it's a vacancy and extracts: title, company, salary, skills, grade, etc.
    *   **Offer Generation**: Generates a personalized cold message based on your CV (provided in JSON) and the vacancy description.
4.  **PocketBase**: Serves as the admin UI, database (SQLite), and API for the frontend.

## Quick Start

### 1. Environment Setup
Create a `.env` file in the project root (next to `pb/`):
```env
TG_API_ID=your_id
TG_API_HASH=your_hash
TG_PHONE=+1234567890
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1 # Optional
```

### 2. Prepare PocketBase
Navigate to `pb/` and start the server:
```bash
go run . serve
```
Access the admin UI (`http://127.0.0.1:8090/_/`) and create an admin account.

### 3. Telegram Authorization
You need to create a session once:
```bash
go run . tg-login
```
Enter the code from Telegram. This generates `session.json` (git-ignored).

### 4. Run
```bash
go run . serve
```

### 5. Run SPA frontend
```bash
pnpm i && pnpm dev
```

## Technologies

*   **Frontend**: Svelte 5 (Runes), Tailwind CSS 4, DaisyUI 5
*   **Backend**: Go + PocketBase
*   **Telegram**: `gotd/td` (MTProto)
*   **AI**: OpenAI API

## Project Structure

- `pb/`: Backend source code (Go).
- `src/`: Frontend source code (SvelteKit).
  - `src/lib/shared`: Generic UI components and PocketBase client.
  - `src/lib/apps`: Feature-specific logic (e.g., `dashboard` for vacancies).
- `session.json`: Telegram session (generated at runtime, git-ignored).
