import { Feedback } from '../../utils/feedback';
import { MANUAL_PREFIX } from './__constants';
import { ManualDoc } from '../../types';
import db from '../../db';

export async function showCommand(name: string) {
  const id = `${MANUAL_PREFIX}${name}`;

  try {
    const manual = (await db.get(id)) as ManualDoc;

    Feedback.text(`📘 Manual: ${name}`);
    Feedback.text(`🕒 Last updated: ${new Date(manual.updatedAt).toLocaleString()}`);
    Feedback.text('──────────────────────────────────────\n');
    Feedback.text(manual.content);
    Feedback.text('\n──────────────────────────────────────');
  } catch {
    Feedback.notFound(`Manual "${name}" not found.`);
  }
}
