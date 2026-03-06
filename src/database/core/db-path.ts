import fs from "node:fs";
import os from "node:os";
import path from "node:path";

export type TartarusEnv = "dev" | "prod";

function getEnvironment(): TartarusEnv {
  const env = process.env.TARTARUS_ENV;
  return env === "dev" ? "dev" : "prod";
}

export function getDatabasePath(): string {
  const env = getEnvironment();

  if (env === "dev") {
    return path.resolve(process.cwd(), "dev.db");
  }

  const prodDirectory = path.resolve(os.homedir(), ".tartarus");
  fs.mkdirSync(prodDirectory, { recursive: true });

  return path.resolve(prodDirectory, "tartarus.db");
}
