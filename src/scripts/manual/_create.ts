import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error, FailThrow } from 'src/utils/error';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { ManualDoc } from 'src/core/types';
import { register } from 'src/utils/register';
import { tempFile } from 'src/scripts/manual/__tempfile';
import db from 'src/core/db';

export const registerManCreate = (program: Command) =>
  register({
    program,
    commandName: 'create',
    commandDescription: 'Create a new manual',
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

        try {
          await db.get(id);
          FailThrow(`Manual "${name}" already exists.`);
        } catch {
          // Not found — ok to proceed
        }

        const manual: ManualDoc = {
          _id: id,
          content: (await tempFile(name, '# New Manual\n')) ?? '',
          updatedAt: new Date().toISOString(),
        };

        await db.put(manual);

        done(`Manual "${name}" was created.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
