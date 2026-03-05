import { initMigration } from "../migrations/001_init";
import type { DatabaseSchema } from "./schema";
import type { Kysely } from "kysely";
import type { Migration } from "../types/migration";

const migrations: Migration[] = [initMigration];

function normalizeVersion(version: string): [number, number, number] {
  const [major = "0", minor = "0", patch = "0"] = version.split(".");
  return [Number(major) || 0, Number(minor) || 0, Number(patch) || 0];
}

function compareVersions(a: string, b: string): number {
  const [aMajor, aMinor, aPatch] = normalizeVersion(a);
  const [bMajor, bMinor, bPatch] = normalizeVersion(b);

  if (aMajor !== bMajor) return aMajor - bMajor;
  if (aMinor !== bMinor) return aMinor - bMinor;
  return aPatch - bPatch;
}

async function ensureMigrationsTable(database: Kysely<DatabaseSchema>): Promise<void> {
  await database.schema
    .createTable("migrations")
    .ifNotExists()
    .addColumn("id", "integer", (col) => col.primaryKey().autoIncrement())
    .addColumn("name", "text", (col) => col.notNull().unique())
    .addColumn("projectVersion", "text", (col) => col.notNull())
    .addColumn("appliedAt", "text", (col) => col.notNull())
    .execute();
}

function getProjectVersion(): string {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const packageJson = require("../../../package.json") as { version?: string };
  return packageJson.version ?? "0.0.0";
}

export async function runMigrator(database: Kysely<DatabaseSchema>): Promise<void> {
  await ensureMigrationsTable(database);

  const projectVersion = getProjectVersion();

  const appliedMigrations = await database
    .selectFrom("migrations")
    .select(["id", "name", "projectVersion", "appliedAt"])
    .orderBy("id", "asc")
    .execute();

  const lastAppliedVersion =
    appliedMigrations.length > 0
      ? appliedMigrations[appliedMigrations.length - 1].projectVersion
      : "0.0.0";

  if (compareVersions(lastAppliedVersion, projectVersion) > 0) {
    const appliedByName = new Map(migrations.map((migration) => [migration.name, migration]));
    const appliedDesc = [...appliedMigrations].reverse();

    for (const applied of appliedDesc) {
      if (compareVersions(applied.projectVersion, projectVersion) <= 0) {
        break;
      }

      const migration = appliedByName.get(applied.name);
      if (!migration) {
        throw new Error(`Migration not found for rollback: ${applied.name}`);
      }

      await migration.down(database);
      await database.deleteFrom("migrations").where("name", "=", migration.name).execute();
    }

    return;
  }

  if (compareVersions(lastAppliedVersion, projectVersion) < 0) {
    const appliedNames = new Set(appliedMigrations.map((migration) => migration.name));
    const sortedMigrations = [...migrations].sort((a, b) =>
      compareVersions(a.projectVersion, b.projectVersion),
    );

    for (const migration of sortedMigrations) {
      if (appliedNames.has(migration.name)) {
        continue;
      }

      if (compareVersions(migration.projectVersion, projectVersion) > 0) {
        continue;
      }

      await migration.up(database);
      await database
        .insertInto("migrations")
        .values({
          name: migration.name,
          projectVersion: migration.projectVersion,
          appliedAt: new Date().toISOString(),
        })
        .execute();
    }
  }
}
