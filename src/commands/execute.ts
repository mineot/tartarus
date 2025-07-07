import db from '@/db';
import { exec } from 'child_process';

export default async function executeCommand(name: string) {
  try {
    const doc = await db.get(name);
    exec(doc.command, (err, stdout, stderr) => {
      if (err) {
        console.error('❌ Error executing command:', err.message);
        return;
      }
      console.log(stdout);
    });
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
