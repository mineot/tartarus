import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { execSync } from 'child_process';
import db from 'src/core/db';

import {
  Args,
  BreakLine,
  command,
  FailThrow,
  ItemText,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerRun = (program: Command) =>
  register({
    program,
    commandName: 'run',
    commandDescription: 'Run command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the command.',
        },
      ],
      example: `tartarus run commandName`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'run',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the command name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          const prefixName = `${COMMAND_PREFIX}${name}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          for (const [index, instruction] of commandDoc.instructions.entries()) {
            ItemText(index, instruction);
            execSync(instruction, { stdio: 'inherit', shell: '/bin/bash' });
            BreakLine();
          }

          return null;
        },
      }),
  });
