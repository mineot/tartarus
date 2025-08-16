import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerAddDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'add',
    commandDescription: 'Add a documentation to a command',
    commandArguments: [
      {
        name: 'name',
        description: 'The name of the command to which the documentation will be added.',
        required: true,
      },
      {
        name: 'doc',
        description: 'The documentation text',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name, doc } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const commandDoc = (await db.get(prefixName)) as CommandDoc;

        commandDoc.description = doc;
        await db.put(commandDoc);

        done(`Documentation "${doc}" has been added to the command "${name}".`);
      } catch (err: any) {
        error(err);
      }
    },
  });
