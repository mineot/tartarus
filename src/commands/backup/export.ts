import db from '../../db';
import fs from 'fs';
import path from 'path';

export default async function exportCommand(filePath: string) {
  const result = await db.allDocs({ include_docs: true });
  const commands = result.rows.map((row) => row.doc).filter(Boolean);

  const resolvedPath = path.resolve(filePath);
  fs.mkdirSync(path.dirname(resolvedPath), { recursive: true });
  fs.writeFileSync(resolvedPath, JSON.stringify(commands, null, 2));

  console.log(`✅ Backup exported to: ${resolvedPath}`);
}
