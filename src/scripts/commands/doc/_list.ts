import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import db from 'src/core/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  TitledText,
  ValidationReturn,
} from 'src/utils/command';

export const registerListDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List commands documentation',
    commandHelp: {
      structure: [
        { name: 'No arguments required', description: 'List all commands documentation' },
      ],
      example: `tartarus cmd doc list`,
    },
    commandInstance: async (args: Args) =>
      command(args, {
        referenceName: 'cmd doc list',
        noArguments: true,
        validation: async (args: Args): ValidationReturn => {
          if (args.length > 0) {
            FailThrow('No arguments required.');
          }
        },
        operation: async (): OperationReturn => {
          const allDocs = await db.allDocs({ include_docs: true });

          const described = allDocs.rows
            .filter((row) => row.id.startsWith(COMMAND_PREFIX))
            .map((row) => row.doc)
            .filter((doc) => (doc as CommandDoc)?.description);

          if (described.length === 0) {
            FailThrow('No documented commands found.');
            return null;
          }

          for (const doc of described) {
            TitledText(
              `"${doc?._id.replace(COMMAND_PREFIX, '')}" documentation`,
              `${(doc as CommandDoc)?.description}`
            );
          }

          return null;
        },
      }),
  });
