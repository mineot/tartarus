import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import db from 'src/core/db';

import {
  Args,
  command,
  OperationReturn,
  register,
  FailThrow,
  ValidationReturn,
} from 'src/utils/command';

export const registerRemoveDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'remove',
    commandDescription: 'Remove a documentation from a command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'The name of the command to which the documentation will be removed.',
        },
      ],
      example: `tartarus cmd doc remove commandName`,
    },
    commandInstance: async (args: Args) =>
      command(args, {
        referenceName: 'cmd doc remove',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the command name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [commandName] = args;
          const prefixName = `${COMMAND_PREFIX}${commandName}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          if (!commandDoc.description) {
            FailThrow(`No documentation to remove from "${commandName}" command.`);
          }

          delete commandDoc.description;

          await db.put(commandDoc);

          return `Documentation has been removed from the command "${commandName}".`;
        },
      }),
  });
