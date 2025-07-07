import db from '@/db';
import fs from 'fs';
import path from 'path';

export default async function exportCommands() {
  const result = await db.allDocs({ include_docs: true });
  const commands = result.rows.map((row) => row.doc).filter(Boolean);
  const filePath = path.join(__dirname, '..', '..', 'backup', 'commands_backup.json');

  fs.mkdirSync(path.dirname(filePath), { recursive: true });
  fs.writeFileSync(filePath, JSON.stringify(commands, null, 2));
  console.log('✅ Backup exported to ${filePath}');
}
