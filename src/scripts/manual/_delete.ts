import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerManDelete = (program: Command) =>
  register({
    program,
    commandName: 'delete',
    commandDescription: 'Delete a manual',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the manual.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name } = args;
        const id = `${MANUAL_PREFIX}${name}`;
        const manual = await db.get(id);
        await db.remove(manual);
        done(`Manual "${name}" was deleted.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
