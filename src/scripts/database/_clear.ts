import db from '../../db';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';

/**
 * Clears all commands from the database.
 * Fetches all documents, marks them for deletion, and then performs a bulk delete operation.
 * Provides feedback on the number of commands deleted or if the database was already empty.
 */
export async function clearCommand() {
  try {
    // Retrieve all documents from the database including their content.
    const allDocs = await db.allDocs({ include_docs: true });

    // Filter and map the documents to mark them for deletion.
    const deletableDocs = allDocs.rows
      .map((row) => row.doc)
      .filter(
        (
          doc
        ): doc is PouchDB.Core.IdMeta &
          PouchDB.Core.RevisionIdMeta &
          CommandDoc & { _deleted?: boolean } => {
          // Ensure the document has valid _id and _rev properties.
          return doc !== undefined && typeof doc._id === 'string' && typeof doc._rev === 'string';
        }
      )
      .map((doc) => ({ ...doc, _deleted: true }));

    // Check if there are no documents to delete.
    if (deletableDocs.length === 0) {
      Feedback.info('No commands to delete. The database is empty.');
      return;
    }

    // Perform a bulk delete operation on the database.
    await db.bulkDocs(deletableDocs);
    Feedback.success(`Successfully deleted ${deletableDocs.length} command(s) from the database.`);
  } catch (error: any) {
    // Log an error message if the operation fails.
    Feedback.error(`Error clearing the database: ${error.message}`);
  }
}
