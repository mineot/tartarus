import db from '@/db';

export default async function listCommands() {
  const result = await db.allDocs({ include_docs: true });

  if (!result.rows.length) {
    console.log('📭 Nenhum comando registrado.');
    return;
  }

  console.log('📚 Comandos registrados:');

  result.rows.forEach((row) => {
    if (row.doc) console.log(`- ${row.id}: ${row.doc.command}`);
  });
}
