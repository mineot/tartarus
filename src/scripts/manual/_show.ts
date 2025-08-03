import { Command } from 'commander';
import { MANUAL_PREFIX } from 'src/utils/constants';
import { ManualDoc } from 'src/types';
import db from 'src/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  TitledText,
  ValidationReturn,
} from 'src/utils/command';

export const registerManShow = (program: Command) =>
  register({
    program,
    commandName: 'show',
    commandDescription: 'Show a manual',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the manual.',
        },
      ],
      example: `tartarus man show manualName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man show',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the manual name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const id = `${MANUAL_PREFIX}${name}`;
          const manual = (await db.get(id)) as ManualDoc;

          TitledText('📘 Manual', name);
          TitledText('🕒 Last updated', new Date(manual.updatedAt).toLocaleString());
          console.log(manual.content);

          return null;
        },
      }),
  });
