# 🧠 Tartarus

**Tartarus** is a CLI tool to manage and run named sets of terminal commands, supporting documentation, manuals, exports, and sequences of instructions.

---

## 🚀 Installation

```bash
npm install -g @mineot/tartarus
```

---

## ⚙️ Requirements

- **Minimum Node.js version:** `18.x`
- [Download Node.js](https://nodejs.org/en/download)

---

## 📦 Command Groups

Tartarus commands are organized into groups for clarity and scalability:

---

### `>_` `cmd` – Command Management

Manage command sets that contain one or more instructions.

```bash
# Create a new command with a single instruction.
tartarus cmd create test "mkdir folder"
```

```bash
# Append a new instruction to an existing command.
tartarus cmd append test "cd folder"
```

```bash
# Subtract an instruction from a command by its index (starting at 0).
tartarus cmd subtract test 0
```

```bash
# Delete a command entirely from the database.
tartarus cmd delete test
```

```bash
# Display all instructions in a specific command.
tartarus cmd show test
```

```bash
# List all registered command names and their first instruction.
tartarus cmd list
```

---

#### 📝 `doc` – Documentation Management

Manage optional documentation descriptions for commands.

```bash
# Add or update a description for a command.
tartarus cmd doc add test "Creates a folder and navigates into it."
```

```bash
# Remove the description from a command.
tartarus cmd doc remove test
```

```bash
# Show documentation for a specific command.
tartarus cmd doc show test
```

```bash
# List all documented commands.
tartarus cmd doc list
```

---

### 🏎️ `run` – Run Command

```bash
# Execute all instructions of the specified command **in order**.
tartarus run test
```

---

### 📘 `man` – Manual Management

Manage textual manuals (long-form documentation) for any command or concept.

```bash
# Sets the default text editor that Tartarus will use for editing manuals. This must be set before using commands like create or edit.
tartarus man set_editor nano # or notepad
```

```bash
# Show the currently configured editor.
tartarus man show_editor
```

```bash
# Create a new manual. Opens the configured editor (e.g. nano, notepad, vim).
tartarus man create deploy-guide
```

```bash
# Edit an existing manual in the configured editor.
tartarus man edit deploy-guide
```

```bash
# Delete a manual from the database.
tartarus man delete deploy-guide
```

```bash
# Display the content of a manual in the terminal.
tartarus man show deploy-guide
```

```bash
# List all existing manual names.
tartarus man list
```

---

### 💾 `db` - Database Management

Manage data persistence and recovery.

```bash
# Export all registered commands to a JSON file.
tartarus db export ~/my_backups/commands_001.json
```

```bash
# Import commands from a JSON file.
tartarus db import ~/my_backups/commands_001.json
```

```bash
# Delete all registered commands from the local database.
# ⚠️ This action is irreversible.
tartarus db clear
```

#### 📁 Backups

- When you export your commands, a `.json` file is created at the path you specify.
- This file contains all registered commands and can be imported later using `tartarus db import`.

---

## 🧱 Storage

Tartarus stores all your commands locally using [PouchDB](https://pouchdb.com/), in a consistent and isolated database.

### 🗃️ Database location by OS

| Operating System | Database Path                              |
| ---------------- | ------------------------------------------ |
| Linux            | `~/.tartarus/db/commands`                  |
| macOS            | `/Users/your-user/.tartarus/db/commands`   |
| Windows          | `C:\Users\your-user\.tartarus\db\commands` |

- This ensures the database is user-specific, secure, and easy to back up.
- You can inspect or remove the database manually if needed.

---

## 🖥️ Local Installation

```bash
git clone https://github.com/mineot/tartarus.git
cd tartarus
npm install
npm run build
npm link
```

> The `npm link` command allows you to use `tartarus` globally on your system.

---

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
