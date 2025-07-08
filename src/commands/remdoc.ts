import db from '../db';

export default async function remDoc(name: string) {
  try {
    const command = await db.get(name);
    if (!command.description) {
      console.log(`ℹ️ No description to remove from "${name}".`);
      return;
    }
    delete command.description;
    await db.put(command);
    console.log(`✅ Description removed from "${name}".`);
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
