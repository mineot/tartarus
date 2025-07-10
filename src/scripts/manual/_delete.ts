import db from '../../db';
import { MANUAL_PREFIX } from './__constants';

export async function deleteCommand(name: string) {
  const id = `${MANUAL_PREFIX}${name}`;

  try {
    const manual = await db.get(id);
    await db.remove(manual);
    console.log(`🗑️ Manual "${name}" deleted.`);
  } catch {
    console.error(`❌ Manual "${name}" not found.`);
  }
}
