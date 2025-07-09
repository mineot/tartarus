import db from '../../db';

export default async function addDoc(name: string, doc: string) {
  try {
    const command = await db.get(name);
    command.description = doc;
    await db.put(command);
    console.log(`✅ Description added to "${name}".`);
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
