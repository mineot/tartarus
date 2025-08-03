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

export const registerManCreate = (program: Command) =>
  register({
    program,
    commandName: 'create',
    commandDescription: 'Create a new manual',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the manual.',
        },
      ],
      example: `tartarus man create manualName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man create',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the manual name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const id = `${MANUAL_PREFIX}${name}`;

          try {
            await db.get(id);
            FailThrow(`Manual "${name}" already exists.`);
            return null;
          } catch {
            // Not found — ok to proceed
          }

          const manual: ManualDoc = {
            _id: id,
            content: (await tempFile(name, '# New Manual\n')) ?? '',
            updatedAt: new Date().toISOString(),
          };

          await db.put(manual);

          return `Manual "${name}" was created.`;
        },
      }),
  });
