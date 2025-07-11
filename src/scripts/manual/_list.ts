import { Feedback } from '../../utils/feedback';
import { MANUAL_PREFIX } from './__constants';
import db from '../../db';

export async function listCommand() {
  try {
    const result = await db.allDocs({ include_docs: false });

    const manuals = result.rows
      .filter((row) => row.id.startsWith(MANUAL_PREFIX))
      .map((row) => row.id.replace(MANUAL_PREFIX, ''));

    if (manuals.length === 0) {
      Feedback.notFound('No manuals found.');
      return;
    }

    Feedback.title('📚 Available Manuals:\n');
    manuals.forEach((name) => Feedback.item(`${name}`));
  } catch (error: any) {
    Feedback.error(`Failed to list manuals: ${error.message}`);
  }
}
