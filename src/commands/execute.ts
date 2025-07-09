import db from '../db';
import { execSync } from 'child_process';

// TODO: AGUARDAR A EXECUÇÃO DE UM COMANDO ANTES DE IR PARA O OUTRO
export default async function execCommand(name: string) {
  try {
    const doc = await db.get(name);
    for (const cmd of doc.instructions) {
      console.log(`➡️  Executing: ${cmd}`);
      const output = execSync(cmd, { stdio: 'inherit', shell: '/bin/bash' });
      console.log(`⬅️  Output: ${output}`);
    }
  } catch {
    console.error(`❌ Command "${name}" not found or execution failed.`);
  }
}
