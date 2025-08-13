import { Command } from 'commander';
import db from 'src/core/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerDbClear = (program: Command) =>
  register({
    program,
    commandName: 'clear',
    commandDescription: 'Clear the database.',
    commandHelp: {
      structure: [],
      example: `tartarus db clear`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'db clear',
        noArguments: true,
        validation: async (args: Args): ValidationReturn => {
          if (args.length > 0) {
            FailThrow('No arguments required.');
          }
        },
        operation: async (): OperationReturn => {
          const result = await db.allDocs({ include_docs: true });

          const deletable = result.rows
            .map((row) => row.doc)
            .filter(
              (doc): doc is PouchDB.Core.ExistingDocument<any> =>
                !!doc && typeof doc._id === 'string' && typeof doc._rev === 'string'
            )
            .map((doc) => ({ ...doc, _deleted: true }));

          await db.bulkDocs(deletable);
          return `Database was cleared.`;
        },
      }),
  });
