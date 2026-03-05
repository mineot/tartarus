import type { Migration } from "../types/migration";

export const initMigration: Migration = {
  name: "001_init",
  projectVersion: "2.0.0",
  async up() {
    // Migration intentionally empty for now.
  },
  async down() {
    // Migration intentionally empty for now.
  },
};
