import { Feedback } from '../../utils/feedback';
import { MANUAL_PREFIX } from './__constants';
import { ManualDoc } from '../../types';
import { tempFile } from './__temp_file';
import db from '../../db';

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
