# 🧠 Tartarus

**Tartarus** é uma ferramenta de linha de comando (CLI) desenvolvida em TypeScript para registrar, listar, executar e manter comandos personalizados do terminal de forma persistente. Os comandos são armazenados localmente usando o banco de dados [PouchDB](https://pouchdb.com/).

---

## 🚀 Instalação

Clone o repositório:

```bash
git clone https://github.com/mineot/tartarus.git
cd tartarus
npm install
npm run build
npm link
```

> O `npm link` permite usar o comando `tartarus` globalmente no seu sistema.

---

## 📦 Comandos disponíveis

### `register <nome> <comando>`

Registra um novo comando:

```bash
tartarus register list_home "ls -ls ~"
```

---

### `exec <nome>`

Executa o comando salvo:

```bash
tartarus exec list_home
```

---

### `list`

Lista todos os comandos registrados:

```bash
tartarus list
```

---

### `remove <nome>`

Remove um comando salvo:

```bash
tartarus remove list_home
```

---

### `export`

Exporta todos os comandos registrados para um arquivo JSON na pasta `backup/`:

```bash
tartarus export
```

---

### `import`

Importa comandos de um arquivo `commands_backup.json` dentro da pasta `backup/`:

```bash
tartarus import
```

---

## 💾 Armazenamento

Os comandos são armazenados localmente com [PouchDB](https://pouchdb.com/), usando arquivos no diretório `db/`.

O backup é salvo em `backup/commands_backup.json`.

---

## 🧑‍💻 Tecnologias utilizadas

- [TypeScript](https://www.typescriptlang.org/)
- [Commander.js](https://github.com/tj/commander.js/)
- [PouchDB](https://pouchdb.com/)
- [Node.js](https://nodejs.org/)

---

## 📄 Licença

Este projeto está licenciado sob a [Licença Apache 2.0](LICENSE).
