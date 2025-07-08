import db from '../db';

export default async function addInstruction(name: string, newInstruction: string) {
  try {
    const doc = await db.get(name);
    doc.instructions.push(newInstruction);
    await db.put(doc);
    console.log(`✅ Instruction added to "${name}".`);
  } catch {
    console.error(`❌ Command "${name}" not found.`);
  }
}
