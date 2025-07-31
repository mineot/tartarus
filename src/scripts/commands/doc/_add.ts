import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/types';
import { register, command, OperatorReturn, Args } from 'src/utils/command';
import db from 'src/db';

export const registerAddDocumentationCommand = (program: Command) =>
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
        operation: async (args: Args): OperatorReturn => {
          const [commandName, descriptionText] = args;
          const prefixName = `${COMMAND_PREFIX}${commandName}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          commandDoc.description = descriptionText;
          await db.put(commandDoc);

          return null;
        },
      }),
  });
