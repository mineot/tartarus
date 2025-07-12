import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

export async function clearCommand() {
  try {
    const result = await db.allDocs({ include_docs: true });

    const deletable = result.rows
      .map((row) => row.doc)
      .filter(
        (doc): doc is PouchDB.Core.ExistingDocument<any> =>
          !!doc && typeof doc._id === 'string' && typeof doc._rev === 'string'
      )
      .map((doc) => ({ ...doc, _deleted: true }));

    if (deletable.length === 0) {
      Feedback.notFound('No documents to delete.');
      return;
    }

    await db.bulkDocs(deletable);
    Feedback.success(`Cleared ${deletable.length} documents from the database.`);
  } catch (error: any) {
    Feedback.error(`Failed to clear database: ${error.message}`);
  }
}
