import { Command } from 'commander';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { ManualDoc } from 'src/types';
import { tempFile } from 'src/scripts/manual/__tempfile';
import db from 'src/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerManEdit = (program: Command) =>
  register({
    program,
    commandName: 'edit',
    commandDescription: 'Edit a manual',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the manual.',
        },
      ],
      example: `tartarus man edit manualName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man edit',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the manual name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const id = `${MANUAL_PREFIX}${name}`;
          let manual: ManualDoc = (await db.get(id)) as ManualDoc;
          manual.content = (await tempFile(name, manual.content)) ?? '';
          manual.updatedAt = new Date().toISOString();
          await db.put(manual);
          return `Manual "${name}" was edited.`;
        },
      }),
  });
