import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { ManualDoc } from 'src/core/types';
import { register } from 'src/utils/register';
import { tempFile } from 'src/scripts/manual/__tempfile';
import db from 'src/core/db';

export const registerManEdit = (program: Command) =>
  register({
    program,
    commandName: 'edit',
    commandDescription: 'Edit an existing manual',
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
        let manual: ManualDoc = (await db.get(id)) as ManualDoc;
        manual.content = (await tempFile(name, manual.content)) ?? '';
        manual.updatedAt = new Date().toISOString();
        await db.put(manual);
        done(`Manual "${name}" was edited.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
