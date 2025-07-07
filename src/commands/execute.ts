import db from '@/db';
import { exec } from 'child_process';

export default async function execute(name: string) {
  try {
    const doc = await db.get(name);
    exec(doc.command, (err, stdout, stderr) => {
      if (err) {
        console.error('❌ Erro na execução:', err.message);
        return;
      }
      console.log(stdout);
    });
  } catch {
    console.error(`❌ Comando "${name}" não encontrado.`);
  }
}
