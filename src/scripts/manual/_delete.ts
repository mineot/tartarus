import { Command } from 'commander';
import { MANUAL_PREFIX } from 'src/utils/constants';
import db from 'src/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerManDelete = (program: Command) =>
  register({
    program,
    commandName: 'delete',
    commandDescription: 'Delete a manual',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the manual.',
        },
      ],
      example: `tartarus man delete manualName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man delete',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the manual name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const id = `${MANUAL_PREFIX}${name}`;
          const manual = await db.get(id);
          await db.remove(manual);
          return `Manual "${name}" was deleted.`;
        },
      }),
  });
