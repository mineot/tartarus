# 🧠 Tartarus

**Tartarus** is a powerful command-line tool written in TypeScript for organizing, documenting, executing, exporting, and importing named command sets. All commands are stored locally using [PouchDB](https://pouchdb.com/), and can be grouped and extended with multiple instructions.

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

### 📁 `cmd` – Command Management

Manage command sets that contain one or more instructions.

#### `tartarus cmd create <name> <command>`

Create a new command with a single instruction.

```bash
tartarus cmd create test "mkdir folder"
```

#### `tartarus cmd append <name> <instruction>`

Append a new instruction to an existing command.

```bash
tartarus cmd append test "cd folder"
```

#### `tartarus cmd subtract <name> <index>`

Subtract an instruction from a command by its index (starting at 0).

```bash
tartarus cmd subtract test 0
```

#### `tartarus cmd remove <name>`

Remove a command entirely from the database.

```bash
tartarus cmd remove test
```

#### `tartarus cmd show <name>`

Display all instructions in a specific command.

```bash
tartarus cmd show test
```

#### `tartarus cmd list`

List all registered command names and their first instruction.

```bash
tartarus cmd list
```

#### `tartarus cmd exec <name>`

Execute all instructions of the specified command **in order**.

```bash
tartarus cmd exec test
```

---

### 📝 `doc` – Documentation Management

Manage optional descriptions for commands.

#### `tartarus doc add <name> <text>`

Add or update a description for a command.

```bash
tartarus doc add test "Creates a folder and navigates into it."
```

#### `tartarus doc remove <name>`

Remove the description from a command.

```bash
tartarus doc remove test
```

#### `tartarus doc show [name]`

Show documentation for a specific command or all documented commands.

```bash
tartarus doc show test
tartarus doc show
```

---

### 💾 `database` – Backup and Maintenance

Manage data persistence and recovery.

#### `tartarus db export <file_path>`

Export all registered commands to a JSON file.

```bash
tartarus db export ~/my_backups/commands_001.json
```

#### `tartarus db import <file_path>`

Import commands from a JSON file.

```bash
tartarus db import ~/my_backups/commands_001.json
```

✅ **Note:** Ensure `_rev` fields are removed from the JSON before importing.

#### `tartarus db clear`

Delete all registered commands from the local database.

```bash
tartarus db clear
```

⚠️ This action is irreversible.

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

### 📁 Backups

- When you export your commands, a `.json` file is created at the path you specify.
- This file contains all registered commands and can be imported later using `tartarus db import`.

## 🖥️ Local Installation

```bash
git clone https://github.com/mineot/tartarus.git
cd tartarus
npm install
npm run build
npm link
```

> The `npm link` command allows you to use `tartarus` globally on your system.

## 📜 NPM Scripts

| Script            | Description                                                                        |
| ----------------- | ---------------------------------------------------------------------------------- |
| `npm run build`   | Builds the project using esbuild into a single minified file (`dist/tartarus.js`). |
| `npm run dev`     | Runs the CLI directly with `ts-node`.                                              |
| `npm run format`  | Formats the code using Prettier.                                                   |
| `npm run prepare` | Automatically builds before linking or publishing.                                 |

---

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
