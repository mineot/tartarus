import db from '../../db';
import { CommandDoc } from '../../types';

export default async function clearCommand() {
  try {
    const result = await db.allDocs({ include_docs: true });

    const docsToDelete = result.rows
      .map((row) => row.doc)
      .filter(
        (
          doc
        ): doc is PouchDB.Core.IdMeta &
          PouchDB.Core.RevisionIdMeta &
          CommandDoc & { _deleted?: boolean } => {
          return !!doc && typeof doc._id === 'string' && typeof doc._rev === 'string';
        }
      )
      .map((doc) => ({ ...doc, _deleted: true }));

    if (docsToDelete.length === 0) {
      console.log('ℹ️ The database is already empty.');
      return;
    }

    await db.bulkDocs(docsToDelete);
    console.log(`✅ ${docsToDelete.length} command(s) deleted from the database.`);
  } catch (error) {
    console.error('❌ Failed to clear the database:', (error as Error).message);
  }
}
