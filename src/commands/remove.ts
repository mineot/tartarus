import db from '../db';

export default async function removeCommand(name: string) {
  try {
    const doc = await db.get(name);
    await db.remove(doc);
    console.log(`✅ Command "${name}" removed.`);
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
