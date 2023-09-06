# Project plan


## TODO List

### Telegram chatbot

- [x] Configure supported
  commands: `/hello`, `gsheets`, `/splitbill`, `/housework`, `/settings`, `/feedback`, `/help`, ...
- [x] Add bot into group
- [x] Configure API token

### Google sheets

- [x] Configure Google Sheets API
- [x] Configure Google Sheets credentials

### Requirements

**Handle all commands when received**

- [ ] reply to user if command is supported
- [ ] reply to user if command is not supported

**Command: `/hello` handler**

- [ ] Greet user

**Command: `/bill` handler**

- [ ] show the list of buttons for bill management: `add`, `view`, `update`, `delete`, `report`
  - [ ] handle `add` button:
    - user input: each on a new line: `name`, `amount`, `date`, `payer`
      ```
      [expense name]: default: name - current date
      [amount]: support parse "k" -> thousand, "m" -> million
      [date]: default: current date
      [payer]: default: current user
      ```
    - add new record to Google Sheets
    - reply to user:
        ```
            Status: <status>
            --- <show data as a row of table> ---
            ID: <id> 
            Expense name: <name>
            Amount: <amount>
            Date: <date>
            Payer: <payer>
        ```
  - [ ] handle `view` button:
    - show last 5 records as table
      ```markdown
            | ID | Expense name | Amount | Date | Payer |
            |:---|:-------------|:-------|:-----|:------|
            | 1  | ...          | ...    | ...  | ...   |
            | 2  | ...          | ...    | ...  | ...   |
        ```
    - show buttons: `next`, `previous`, `back` (optional)
  - [ ] handle `report` button:
    - show report as table
      ```markdown
        |                               | Amount    | ***        |
        |:------------------------------|:----------|:-----------|
        | living expenses               | 5.000.000 | 2.500.000  |
        | person1 paid                  | 1.000.000 | 1.500.000  |
        | person2 paid                  | 4.000.000 | -1.500.000 |
        | rent                          | 4.500.000 | 2.250.000  |
        | total                         | 9.500.000 | 4.750.000  |
        | gap= (total)/2-[person2 paid] | 3.250.000 |            |
      ``` 

**Command: `/housework` handler**

- Show the list of buttons for housework management: `list`, `add`, `update`, `delete`
  - [ ] handle `list` button:
    - show the list of housework tasks as table
      ```markdown
        | ID | Task name | Frequency | Last done | Next due | Next assignee |
        |:---|:----------|:----------|:----------|:---------|:--------------|
        | 1  | ...       | ...       | ...       | ...      |...            |
        | 2  | ...       | ...       | ...       | ...      |...            |
      ```
    - show buttons: each task as a button, `back` (optional)
    - [ ] handle `task selected` button
      - show task details
        ```
            ID: <id> 
            Task name: <name>
            Frequency: <frequency>
            Last done: <last done>
            Next due: <next due>
            Next assignee: <next assignee>
        ```
    - show buttons: `mark as done`, `remind housemates`, `back` (optional)
  - [ ] handle `add` button:
    - user input: each on a new line: `name`, `frequency`, `last done`
      ```
      [task name]: default: name - current date
      [frequency]: default: 1 week
      [next due]: default: current date
      [next assignee]: default: current user
      ```
    - add new record to Google Sheets
    - reply to user:
        ```
            Status: <status>
            --- <show data as a row of table> ---
            ID: <id> 
            Task name: <name>
            Frequency: <frequency>
            Last done: <last done>
            Next due: <next due>
            Next assignee: <next assignee>
        ```

**Command: `/sgheets` handler**

- [ ] show buttons: `list`, `create`, `select main sheet`, `back` (optional)
- [ ] handle `list` button:
  - show the list of sheets as table
    ```markdown
      | ID | Sheet name | Sheet ID |
      |:---|:-----------|:---------|
      | 1  | ...        | ...      |
      | 2  | ...        | ...      |
    ```
  - show buttons: `select main sheet`, `back` (optional)
- [ ] handle `create` button:
  - create new sheet with name: `Housematee - <current month>/<current year>`
  - add new record to Google Sheets
  - reply to user:
      ```
          Status: <status>
          --- <show data as a row of table> ---
          ID: <id> 
          Sheet name: <name>
          Sheet ID: <sheet id>
      ```
- [ ] handle `select main sheet` button:

**Command: `/help` handler**

- [ ] show the list of buttons for
  help: `/hello`, `gsheets`, `/splitbill`, `/housework`, `/settings`, `/feedback`, `/help`, ...

----------------

## Phase 1: MVP
- Create a Telegram bot that can:
    - Show a list of available commands
    - Interact with users via commands
    - Response mock data (including the message, buttons, tables, etc.)

## Phase 2: Data persistence
- Store data in a database (google sheets)
- Implement must-have features:
    - Split bill
    - Housework
    - Settings
    - ...

## Phase 3: Enhance user experience
- Implement nice-to-have features:
    
---
