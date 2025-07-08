import db from '../../db';

export default async function listCommands() {
  const result = await db.allDocs({ include_docs: true });
  if (!result.rows.length) {
    console.log('📭 No commands registered.');
    return;
  }

  console.log('📚 Registered commands:');
  for (const row of result.rows) {
    const doc = row.doc;
    if (doc) {
      console.log(`\n🔹 ${row.id}`);
      doc.instructions.forEach((cmd, index) => {
        console.log(`   ${index + 1}. ${cmd}`);
      });
    }
  }
}
