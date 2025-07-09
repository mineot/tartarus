# 🧠 Tartarus

**Tartarus** is a command-line tool (CLI) written in TypeScript that allows you to register, list, execute, and manage custom terminal commands persistently. All commands are stored locally using [PouchDB](https://pouchdb.com/).

---

## 🚀 Installation

Clone the repository:

```bash
git clone https://github.com/mineot/tartarus.git
cd tartarus
npm install
npm run build
npm link
```

> The `npm link` command allows you to use `tartarus` globally on your system.

---

## ⚙️ Requirements

- **Minimum Node.js version:** `18.x`
- [Download Node.js](https://nodejs.org/en/download)

---

## 📦 Available Commands

### `register <name> <command>`

Register a new command:

```bash
tartarus register list_home "ls -ls ~"
```

---

### `exec <name>`

Execute all instructions registered under a command (in sequence):

```bash
tartarus exec list_home
```

This runs each instruction in the order they were added.

---

### `list`

List all registered commands:

```bash
tartarus list
```

---

### `remove <name>`

Remove a registered command:

```bash
tartarus remove list_home
```

---

### `export <file_path>`

Export all registered commands to a specific file path:

```bash
tartarus export ~/my_backups/commands_001.json
```

This will write the JSON backup to the exact location you specify. The directory will be created if it does not exist.

---

### `import <file_path>`

Import commands from a specific JSON file path:

```bash
tartarus import ~/my_backups/commands_001.json
```

This will read the commands from the given path and store them in the local database.

---

### `clear`

Remove all registered commands from the local database:

```bash
tartarus clear
```

⚠️ This action is irreversible. Use with caution.

---


---

### `adddoc <name> <text>`

Attach a documentation string to a command:

```bash
tartarus adddoc test "Creates a folder and navigates into it."
```

This description will be stored with the command and can be viewed later.

---

### `remdoc <name>`

Remove the documentation from a command:

```bash
tartarus remdoc test
```

---

### `showdoc [name]`

Display documentation for all commands, or a specific one:

```bash
tartarus showdoc
tartarus showdoc test
```

This allows you to understand what each command is intended to do.



---

### `rminst <name> <index>`

Remove an instruction from a command by its index:

```bash
tartarus rminst test 0
```

This will remove the instruction at position 0 from the `test` command. Indexes start from 0.


## 💾 Storage

Commands are stored locally using [PouchDB](https://pouchdb.com/), in the `db/` directory.

Backups are saved in `backup/commands_backup.json`.

---

## 📜 NPM Scripts

These development scripts help you build, test, and manage the CLI tool:

| Script            | Description                                                                                  |
| ----------------- | -------------------------------------------------------------------------------------------- |
| `npm run build`   | Compiles the TypeScript code using esbuild into a single minified file (`dist/tartarus.js`). |
| `npm run dev`     | Runs the CLI directly using ts-node (useful during development).                             |
| `npm run format`  | Formats all `.ts` files in `src/` using Prettier.                                            |
| `npm run prepare` | Automatically runs `build` when installing or linking this package.                          |
| `npm run link`    | Builds and links the `tartarus` CLI globally on your system.                                 |
| `npm run unlink`  | Unlinks the globally installed `tartarus` command.                                           |
| `npm run update`  | Rebuilds and relinks the CLI — useful after making changes.                                  |

---

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).

---

### `add <name> <instruction>`

Add a new instruction to an existing registered command:

```bash
tartarus add test "cd folder"
```

You can register a command with one instruction using `register`, and then keep appending more steps using `add`. When you run `exec`, all instructions will be executed **in order**.
