import { Kysely, SqliteDialect } from "kysely";
import { runMigrator } from "./core/migrator";
import path from "node:path";
import SQLite from "better-sqlite3";
import type { DatabaseSchema } from "./core/schema";

const databaseFilePath = path.resolve(process.cwd(), "tartarus.db");

export const database = new Kysely<DatabaseSchema>({
  dialect: new SqliteDialect({
    database: new SQLite(databaseFilePath),
  }),
});

export const databaseReady = runMigrator(database);
