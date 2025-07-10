import { Feedback } from '../../utils/feedback';
import { MANUAL_PREFIX } from './__constants';
import { ManualDoc } from '../../types';
import { tempFile } from './__temp_file';
import db from '../../db';

export async function editCommand(name: string) {
  const id = `${MANUAL_PREFIX}${name}`;

  let manual: ManualDoc;

  try {
    manual = (await db.get(id)) as ManualDoc;
  } catch {
    Feedback.error(`Manual "${name}" not found.`);
    return;
  }

  manual.content = (await tempFile(name, manual.content)) ?? '';
  manual.updatedAt = new Date().toISOString();

  await db.put(manual);
  Feedback.success(`Manual "${name}" updated.`);
}
