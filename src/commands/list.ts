import db from '@/db';

export default async function listCommand() {
  const result = await db.allDocs({ include_docs: true });

  if (!result.rows.length) {
    console.log('📭 No commands registered.');
    return;
  }

  console.log('📚 Registered commands:');

  result.rows.forEach((row) => {
    if (row.doc) console.log(`- ${row.id}: ${row.doc.command}`);
  });
}
