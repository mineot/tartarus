import db from '../../db';

export default async function showDoc(name?: string) {
  if (name) {
    try {
      const doc = await db.get(name);
      console.log(`📘 ${name}:`);
      console.log(doc.description ?? 'No description found.');
    } catch {
      console.error(`❌ Command "${name}" not found.`);
    }
    return;
  }

  const result = await db.allDocs({ include_docs: true });
  const docs = result.rows.map((r) => r.doc).filter(Boolean);

  const described = docs.filter((doc) => doc?.description);

  if (described.length === 0) {
    console.log('📭 No documented commands found.');
    return;
  }

  for (const doc of described) {
    console.log(`📘 ${doc?._id}:\n${doc?.description}\n`);
  }
}
