import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import { output } from 'src/utils/outputs';
import db from 'src/core/db';

export const registerListDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List all command documentations',
    commandArguments: [],
    commandInstance: async () => {
      try {
        const allDocs = await db.allDocs({ include_docs: true });

        const described = allDocs.rows
          .filter((row) => row.id.startsWith(COMMAND_PREFIX))
          .map((row) => row.doc)
          .filter((doc) => (doc as CommandDoc)?.description);

        if (described.length === 0) {
          FailThrow('No documented commands found.');
          return;
        }

        for (const doc of described) {
          output({
            title: doc?._id.replace(COMMAND_PREFIX, '') ?? '',
            list: [(doc as CommandDoc)?.description ?? ''],
          });
        }
      } catch (err: any) {
        error(err);
      }
    },
  });
