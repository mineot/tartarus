import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
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
} from 'src/utils/command';

export const registerCmdShow = (program: Command) =>
  register({
    program,
    commandName: 'show',
    commandDescription: 'Show command details',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the command.',
        },
      ],
      example: `tartarus cmd show commandName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd show',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the command name to show.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const prefixName = `${COMMAND_PREFIX}${name}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          if (commandDoc.instructions.length === 0) {
            FailThrow(`No instructions available for the command "${name}".`);
            return null;
          }

          TitledText(
            `Command ${name}`,
            commandDoc.description ?? 'Below is the list of instructions'
          );

          commandDoc.instructions.forEach((instruction, index) => {
            ItemText(index, instruction);
          });

          return null;
        },
      }),
  });
