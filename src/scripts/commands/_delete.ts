import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerCmdDelete = (program: Command) =>
  register({
    program,
    commandName: 'delete',
    commandDescription: 'Delete a exists command',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the command.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const command = await db.get(prefixName);

        await db.remove(command);

        done(`Command "${name}" deleted.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
