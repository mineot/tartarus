import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import db from 'src/core/db';

import {
  Args,
  command,
  TitledText,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerShowDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'show',
    commandDescription: 'Show command documentation',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'The name of the command to show documentation for.',
        },
      ],
      example: `tartarus cmd doc show commandName`,
    },
    commandInstance: async (args: Args) =>
      command(args, {
        referenceName: 'cmd doc show',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the command name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [commandName] = args;
          const prefixName = `${COMMAND_PREFIX}${commandName}`;
          const doc = (await db.get(prefixName)) as CommandDoc;

          if (!doc.description) {
            FailThrow('No documentation to show.');
            return null;
          }

          TitledText(`"${commandName}" documentation`, doc.description);

          return null;
        },
      }),
  });
