import db from '../db';
import fs from 'fs';
import path from 'path';
import { CommandDoc } from '../types';

export default async function importCommand(filePath: string) {
  const resolvedPath = path.resolve(filePath);

  if (!fs.existsSync(resolvedPath)) {
    console.error(`❌ File not found: ${resolvedPath}`);
    return;
  }

  const data: CommandDoc[] = JSON.parse(fs.readFileSync(resolvedPath, 'utf-8'));
  await db.bulkDocs(data);

  console.log(`✅ Backup imported from: ${resolvedPath}`);
}
