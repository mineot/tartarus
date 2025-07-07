import db from '@/db';

export default async function removeCommand(name: string) {
  try {
    const doc = await db.get(name);
    await db.remove(doc);
    console.log(`✅ Comando "${name}" removido.`);
  } catch {
    console.error(`❌ Comando "${name}" não encontrado.`);
  }
}
