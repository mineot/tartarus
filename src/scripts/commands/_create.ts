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

export const registerCmdCreate = (program: Command) =>
  register({
    program,
    commandName: 'create',
    commandDescription: 'Create a command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the command.',
        },
        {
          name: 'instruction',
          description: 'Provide the first instruction to add to the command.',
        },
      ],
      example: `tartarus cmd create commandName "firstInstruction"`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd create',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 2) {
            FailThrow(
              'You must provide two arguments: the command name and the first instruction.'
            );
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name, instruction] = args;
          const prefixName = `${COMMAND_PREFIX}${name}`;
          const command = await db.get(prefixName).catch(() => null);

          if (command) {
            FailThrow(`Command "${name}" already exists`);
          }

          await db.put({ _id: prefixName, instructions: [instruction] });

          return `Command "${name}" with instruction "${instruction}" was created.`;
        },
      }),
  });
