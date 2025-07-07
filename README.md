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

### `export`

Export all registered commands into a JSON file inside the `backup/` folder:

```bash
tartarus export
```

---

### `import`

Import commands from `backup/commands_backup.json`:

```bash
tartarus import
```

---

## 💾 Storage

Commands are stored locally using [PouchDB](https://pouchdb.com/), in the `db/` directory.

Backups are saved in `backup/commands_backup.json`.

---

## 🧑‍💻 Technologies Used

- [TypeScript](https://www.typescriptlang.org/)
- [Commander.js](https://github.com/tj/commander.js/)
- [PouchDB](https://pouchdb.com/)
- [Node.js](https://nodejs.org/)

---

## 📄 License

This project is licensed under the [Apache 2.0 License](LICENSE).
