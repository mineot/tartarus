import db from '../../../db';

export default async function removeInstruction(name: string, index: string) {
  try {
    const doc = await db.get(name);
    const idx = parseInt(index, 10);

    if (isNaN(idx) || idx < 0 || idx >= doc.instructions.length) {
      console.error(`❌ Invalid instruction index for "${name}".`);
      return;
    }

    const removed = doc.instructions.splice(idx, 1);
    await db.put(doc);

    console.log(`✅ Removed instruction [${idx}] from "${name}": ${removed[0]}`);
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
