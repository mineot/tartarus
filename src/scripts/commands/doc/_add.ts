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

export const registerAddDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'add',
    commandDescription: 'Add a documentation to a command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'The name of the command to which the documentation will be added.',
        },
        { name: 'doc', description: 'The documentation text' },
      ],
      example: `tartarus cmd doc add commandName "The command documentation"`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd doc add',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 2) {
            FailThrow(
              'You must provide two arguments: the command name and the documentation text.'
            );
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [commandName, descriptionText] = args;
          const prefixName = `${COMMAND_PREFIX}${commandName}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          commandDoc.description = descriptionText;
          await db.put(commandDoc);

          return `Documentation "${descriptionText}" has been added to the command "${commandName}".`;
        },
      }),
  });
