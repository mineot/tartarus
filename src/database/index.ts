import { getDatabasePath } from './core/db-path';
import { Kysely, SqliteDialect } from 'kysely';
import { runMigrator } from './core/migrator';
import SQLite from 'better-sqlite3';
import type { DatabaseSchema } from './core/schema';

export const database = new Kysely<DatabaseSchema>({
  dialect: new SqliteDialect({
    database: new SQLite(getDatabasePath()),
  }),
});

export const databaseMigrator = runMigrator(database);
