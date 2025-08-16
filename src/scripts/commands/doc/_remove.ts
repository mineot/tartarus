import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { done } from 'src/utils/outputs';
import { FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerRemoveDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'remove',
    commandDescription: 'Remove a documentation from a command',
    commandArguments: [
      {
        name: 'name',
        description: 'The name of the command to which the documentation will be removed.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      const { name } = args;
      const prefixName = `${COMMAND_PREFIX}${name}`;
      const commandDoc = (await db.get(prefixName)) as CommandDoc;

      if (!commandDoc.description) {
        FailThrow(`No documentation to remove from "${name}" command.`);
        return;
      }

      delete commandDoc.description;
      await db.put(commandDoc);

      done(`Documentation has been removed from the command "${name}".`);
    },
  });
