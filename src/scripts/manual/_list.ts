import { Feedback } from 'src/utils/feedback';
import { MANUAL_PREFIX } from 'src/scripts/manual/__constants';
import db from 'src/db';

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

    Feedback.title('📚 Available Manuals:');
    manuals.forEach((name) => Feedback.item(`${name}`));
  } catch (error: any) {
    Feedback.error(`Failed to list manuals: ${error.message}`);
  }
}
