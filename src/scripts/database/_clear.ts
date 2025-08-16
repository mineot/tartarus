import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerDbClear = (program: Command) =>
  register({
    program,
    commandName: 'clear',
    commandDescription: 'Clear the database.',
    commandArguments: [],
    commandInstance: async (args: any) => {
      try {
        const result = await db.allDocs({ include_docs: true });

        const deletable = result.rows
          .map((row) => row.doc)
          .filter(
            (doc): doc is PouchDB.Core.ExistingDocument<any> =>
              !!doc && typeof doc._id === 'string' && typeof doc._rev === 'string'
          )
          .map((doc) => ({ ...doc, _deleted: true }));

        await db.bulkDocs(deletable);

        done(`Database was cleared.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
