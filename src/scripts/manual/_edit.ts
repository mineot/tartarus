import { Feedback } from 'src/utils/feedback';
import { MANUAL_PREFIX } from 'src/scripts/manual/__constants';
import { ManualDoc } from 'src/types';
import { tempFile } from 'src/scripts/manual/__temp_file';
import db from 'src/db';

export async function editCommand(name: string) {
  try {
    const id = `${MANUAL_PREFIX}${name}`;
    let manual: ManualDoc = (await db.get(id)) as ManualDoc;
    manual.content = (await tempFile(name, manual.content)) ?? '';
    manual.updatedAt = new Date().toISOString();
    await db.put(manual);
    Feedback.success(`Manual "${name}" updated.`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Manual "${name}" not found.`);
    } else {
      Feedback.error(`Failed to edit manual "${name}": ${error.message}`);
    }
  }
}
