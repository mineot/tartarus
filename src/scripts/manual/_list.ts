import { Command } from 'commander';
import { MANUAL_PREFIX } from 'src/utils/constants';
import db from 'src/core/db';

import {
  Args,
  command,
  FailThrow,
  ItemText,
  OperationReturn,
  register,
  TitledText,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerManList = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List available manuals',
    commandHelp: {
      structure: [],
      example: `tartarus man list`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man list',
        noArguments: true,
        validation: async (args: Args): ValidationReturn => {
          if (args.length > 0) {
            FailThrow('No arguments required.');
          }
        },
        operation: async (): OperationReturn => {
          const result = await db.allDocs({ include_docs: false });

          const manuals = result.rows
            .filter((row) => row.id.startsWith(MANUAL_PREFIX))
            .map((row) => row.id.replace(MANUAL_PREFIX, ''));

          if (manuals.length === 0) {
            FailThrow('No manuals found.');
            return null;
          }

          TitledText('📚 Available Manuals', '');
          manuals.forEach((name, index) => ItemText(index, name));
          return null;
        },
      }),
  });
