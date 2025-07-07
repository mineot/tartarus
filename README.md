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

Execute a registered command:

```bash
tartarus exec list_home
```

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
