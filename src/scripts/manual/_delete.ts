import { Feedback } from 'src/utils/feedback';
import { MANUAL_PREFIX } from 'src/scripts/manual/__constants';
import db from 'src/db';

export async function deleteCommand(name: string) {
  try {
    const id = `${MANUAL_PREFIX}${name}`;
    const manual = await db.get(id);
    await db.remove(manual);

    Feedback.success(`Manual "${name}" deleted.`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Manual "${name}" not found.`);
    } else {
      Feedback.error(`Failed to delete manual "${name}": ${error.message}`);
    }
  }
}
