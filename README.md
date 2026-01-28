# Housematee Telegram Bot

> Share expenses, split rent, and manage housework with your housemates - all through Telegram.

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot-2CA5E0?style=flat&logo=telegram)](https://core.telegram.org/bots)

<table>
  <tr>
    <td><img src="docs/commands.png" alt="Commands" width="280px"></td>
    <td><img src="docs/split_bill.png" alt="Split bill" width="280px"></td>
    <td><img src="docs/housework.png" alt="Housework" width="280px"></td>
  </tr>
</table>

## Why Housematee?

Living with housemates is great, but managing shared expenses and chores can be a headache. Housematee solves this by:

- **No more "who paid for what?"** - Track every expense with full audit history
- **Fair rent splitting** - Weighted utility sharing based on room size or agreement
- **Rotating chores** - Automatic task rotation so everyone does their fair share
- **All in Telegram** - No extra apps needed, works right in your group chat

## Features

### Expense Tracking (`/splitbill`)
- Add expenses with smart parsing (`100k` = 100,000)
- View recent expenses with quick Update/Delete buttons
- Complete audit trail - see who changed what and when
- Monthly reports showing who owes whom

### Rent Management (`/rent`)
- Step-by-step rent entry (total, electric, water)
- **Weighted splitting** - Electric/water split by member weight, other fees split equally
- Per-person breakdown with exact amounts

### Housework Rotation (`/housework`)
- Set up recurring tasks with custom frequencies
- Automatic rotation between housemates
- Daily reminders at 18:30 for due tasks
- Quick shortcuts: `/hw1`, `/hw2` to mark tasks done

### Google Sheets Integration
- All data stored in your own Google Spreadsheet
- Create monthly sheets from template
- Full control over your data

## Quick Start

### 1. Add the Bot to Your Group
Search for your bot on Telegram and add it to your housemates group.

### 2. Set Up Google Sheets
1. Copy the [sample spreadsheet](https://docs.google.com/spreadsheets/d/1a_etCpFf-B1woVM9qjLPM0Nzwox3KPm_ok2bSibdgJk/edit?usp=sharing)
2. Configure Google Sheets API credentials
3. Update the config with your spreadsheet ID

### 3. Start Using!
```
/splitbill     - Manage shared expenses
/rent          - Add monthly rent with breakdown
/housework     - Manage and track chores
/help          - See all commands
```

## Commands

| Command | Description |
|---------|-------------|
| `/splitbill` | Expense management - add, view, update, delete, report |
| `/splitbill_add` | Quick add an expense |
| `/rent` | Add rent with electric/water/other breakdown |
| `/housework` | View and manage household chores |
| `/hw1`, `/hw2` | Quick mark task 1, 2 as done |
| `/gsheets` | Create new monthly sheet |
| `/settings` | Toggle reminders on/off |
| `/help` | Show all available commands |
| `/cancel` | Cancel current operation |

## How It Works

### Adding an Expense
```
/splitbill -> Add -> Enter details:
---
Groceries
150k
25/01/2026
@username
```
The bot will parse `150k` as 150,000 and auto-fill date/payer if not provided.

### Rent Calculation Example
```
Total rent: 5,000,000
Electric: 300,000  (split by weight)
Water: 200,000     (split by weight)
Other: 4,500,000   (split equally)

Member weights: @alice (3), @bob (2)

@alice pays: 180,000 + 120,000 + 2,250,000 = 2,550,000
@bob pays:   120,000 + 80,000 + 2,250,000  = 2,450,000
```

### Audit Trail
Every change is tracked:
```
[25/01/2026 10:30]: amount: 150,000 - by @alice
[25/01/2026 14:15]: update amount: 160,000 - by @bob
```

## Self-Hosting

### Prerequisites
- Go 1.25+
- Google Cloud project with Sheets API enabled
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))

### Installation

```bash
# Clone the repository
git clone https://github.com/tasszz2k/housematee-tgbot.git
cd housematee-tgbot

# Copy and configure
cp config/conf.yaml.sample config/conf.yaml
# Edit config/conf.yaml with your settings

# Build and run
go build -o housematee ./cmd/main.go
./housematee
```

### Configuration

```yaml
telegram:
  token: "YOUR_BOT_TOKEN"
  allowed_channels:
    - -1001234567890  # Your group chat ID

google_sheets:
  spreadsheet_id: "YOUR_SPREADSHEET_ID"
  credentials_file: "config/credentials.json"
```

## Tech Stack

- **Go** - Fast, reliable backend
- **gotgbot/v2** - Telegram Bot API
- **Google Sheets API** - Data storage
- **Viper** - Configuration management
- **Logrus** - Structured logging
- **Cron** - Scheduled reminders

## Contributing

Contributions are welcome! Feel free to:
- Open issues for bugs or feature requests
- Submit pull requests
- Share feedback via `/feedback` command

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Made with love for housemates everywhere.**
