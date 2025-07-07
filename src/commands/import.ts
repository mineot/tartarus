import { CommandDoc } from '@/types';
import db from '@/db';
import fs from 'fs';
import path from 'path';

export default async function importCommands() {
  const filePath = path.join(__dirname, '..', '..', 'backup', 'commands_backup.json');

  if (!fs.existsSync(filePath)) {
    console.error('❌ Arquivo de backup não encontrado.');
    return;
  }

  const data: CommandDoc[] = JSON.parse(fs.readFileSync(filePath, 'utf-8'));
  await db.bulkDocs(data);
  console.log('✅ Backup importado com sucesso.');
}
