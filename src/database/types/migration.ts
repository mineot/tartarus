import type { DatabaseSchema } from "../core/schema";
import type { Kysely } from "kysely";

export type Migration = {
  name: string;
  projectVersion: string;
  up: (database: Kysely<DatabaseSchema>) => Promise<void>;
  down: (database: Kysely<DatabaseSchema>) => Promise<void>;
};
