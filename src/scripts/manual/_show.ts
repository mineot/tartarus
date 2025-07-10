import { MANUAL_PREFIX } from './__constants';
import { ManualDoc } from '../../types';
import db from '../../db';

export async function showCommand(name: string) {
  const id = `${MANUAL_PREFIX}${name}`;

  try {
    const manual = (await db.get(id)) as ManualDoc;

    console.log(`📘 Manual: ${name}`);
    console.log(`🕒 Last updated: ${new Date(manual.updatedAt).toLocaleString()}`);
    console.log('──────────────────────────────────────\n');
    console.log(manual.content);
    console.log('\n──────────────────────────────────────');
  } catch {
    console.error(`❌ Manual "${name}" not found.`);
  }
}
