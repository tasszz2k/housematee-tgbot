# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] - 2026-01-28

### Added

- **Expense Audit Logging**: Complete audit trail for all expense operations
  - Add: `[DD/MM/YYYY HH:mm]: amount: X ₫ - by @username`
  - Update: `[DD/MM/YYYY HH:mm]: update amount: X ₫ - by @username`
  - Delete: `[DD/MM/YYYY HH:mm]: deleted: name - X ₫ - by @username`
  - All entries append to existing Note field

- **Action Buttons**: Update/Delete inline buttons shown after adding or updating an expense
  - Allows quick modification without navigating back to menu

- **Settings Command** (`/settings`): Runtime configuration for bot features
  - Housework Reminders submenu with ON/OFF toggle
  - Thread-safe state management

### Changed

- **Expense Update Flow**: Simplified to amount-only update
  - Shows current expense details
  - Prompts for new amount only (no name/date/payer changes)

- **Expense Delete**: Changed from hard delete to soft delete
  - Keeps expense ID in the row
  - Clears name, amount, date, payer, participants
  - Appends deletion entry to audit log
  - Soft-deleted expenses filtered from view/update/delete lists

- **Weighted Rent Splitting**: Per-member breakdown now uses configured weights
  - Electric/Water split by member weight
  - Other fees split equally
  - Fixed member data reading from correct columns (O:Q, row 4+)

### Fixed

- Fixed housework Note field displaying Channel ID instead of actual note
- Fixed rent calculation showing 50/50 split despite configured weights
- Fixed update expense message using wrong parse mode (html -> markdown)

## [1.2.0] - 2026-01-26

### Added

- **Rent Command** (`/rent`): New multi-step conversation flow for adding rent with detailed breakdown
  - Collects total bill, electric, and water amounts
  - Auto-fills payer with the user who initiated the command
  - Calculates "Other Fees" automatically (total - electric - water)
  - Writes breakdown to Google Sheets cells J5-J8 and M8
  - Per-member shares calculated by Google Sheets formulas

- **GSheets Create Command** (`/gsheets`): Create new monthly sheets from template
  - Shows confirmation with draft sheet name (YYYY_MM format)
  - Copies Template sheet and renames to current month
  - Updates Database!B2 with new sheet name
  - Returns status with sheet ID

- **Weighted Task Rotation** for housework
  - Added "Turns Remaining" column to track consecutive turns per assignee
  - Task Weights configuration (Task ID, Username, Weight) in columns K-M
  - Rotation logic: decrement turns, then rotate to next member with their weight
  - "Assign to Other" button to skip to next member in rotation

### Changed

- Updated Google Sheets template structure:
  - Report section expanded to include Electric, Water, Other Fees rows (I3:M9)
  - Balances section moved to row 13 (previously row 9)
  - Members section supports optional Weight column

- Improved report display with emojis for better readability
- Updated field labels in balance report to match new spreadsheet columns

### Fixed

- Fixed balance reading to correctly parse new spreadsheet layout
- Fixed report parsing for expanded rent section (7 rows instead of 4)

## [1.1.0] - 2025-10-15

### Added

- Housework task management with list, view, mark as done, and remind features
- Split bill management with add, view, and report functionality
- Google Sheets integration for persistent data storage

### Changed

- Improved error handling for Google Sheets API calls

## [1.0.0] - 2025-09-01

### Added

- Initial release
- Basic bot commands: `/hello`, `/help`, `/settings`, `/feedback`
- Google Sheets API integration
- Telegram bot framework with gotgbot

---

## Version History Summary

| Version | Date | Highlights |
|---------|------|------------|
| 1.3.0 | 2026-01-28 | Expense audit logging, Update/Delete flow, Settings command |
| 1.2.0 | 2026-01-26 | Rent command, GSheets create, Weighted task rotation |
| 1.1.0 | 2025-10-15 | Housework and Split bill features |
| 1.0.0 | 2025-09-01 | Initial release |

[Unreleased]: https://github.com/tasszz2k/housematee-tgbot/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/tasszz2k/housematee-tgbot/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/tasszz2k/housematee-tgbot/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/tasszz2k/housematee-tgbot/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/tasszz2k/housematee-tgbot/releases/tag/v1.0.0
