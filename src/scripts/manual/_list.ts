import { Command } from 'commander';
import { error, FailThrow } from 'src/utils/error';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { output } from 'src/utils/outputs';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerManList = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List available manuals',
    commandArguments: [],
    commandInstance: async () => {
      try {
        const result = await db.allDocs({ include_docs: false });

        const manuals = result.rows
          .filter((row) => row.id.startsWith(MANUAL_PREFIX))
          .map((row) => row.id.replace(MANUAL_PREFIX, ''));

        if (manuals.length === 0) {
          FailThrow('No manuals found.');
        }

        output({
          title: 'Available Manuals',
          list: manuals.map((name) => name),
        });
      } catch (err: any) {
        error(err);
      }
    },
  });
