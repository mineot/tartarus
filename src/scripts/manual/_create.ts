import { Feedback } from '../../utils/feedback';
import { MANUAL_PREFIX } from './__constants';
import { ManualDoc } from '../../types';
import { tempFile } from './__temp_file';
import db from '../../db';

export async function createCommand(name: string) {
  try {
    const id = `${MANUAL_PREFIX}${name}`;

    try {
      await db.get(id);
      Feedback.error(`Manual "${name}" already exists.`);
      return;
    } catch {
      // Not found — ok to proceed
    }

    const manual: ManualDoc = {
      _id: id,
      content: (await tempFile(name, '# New Manual\n')) ?? '',
      updatedAt: new Date().toISOString(),
    };

    await db.put(manual);
    Feedback.success(`Manual "${name}" created.`);
  } catch (error: any) {
    Feedback.error(`Failed to create manual "${name}": ${error.message}`);
  }
}
