import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import db from 'src/core/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerCmdDelete = (program: Command) =>
  register({
    program,
    commandName: 'delete',
    commandDescription: 'Delete a command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the command.',
        },
      ],
      example: `tartarus cmd delete commandName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd delete',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the command name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const prefixName = `${COMMAND_PREFIX}${name}`;
          const command = await db.get(prefixName);
          await db.remove(command);
          return `Command "${name}" deleted.`;
        },
      }),
  });
