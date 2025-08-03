import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/types';
import db from 'src/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerCmdAppend = (program: Command) =>
  register({
    program,
    commandName: 'append',
    commandDescription: 'Append instruction to command',
    commandHelp: {
      structure: [
        {
          name: 'name',
          description: 'Provide the name of the command.',
        },
        {
          name: 'instruction',
          description: 'Provide the instruction to add to the command.',
        },
      ],
      example: `tartarus cmd append commandName "newInstruction"`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd append',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 2) {
            FailThrow('You must provide two arguments: the command name and the new instruction.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [name, instruction] = args;
          const prefixName = `${COMMAND_PREFIX}${name}`;
          const commandDoc = (await db.get(prefixName)) as CommandDoc;

          commandDoc.instructions.push(instruction);
          await db.put(commandDoc);

          return `Instruction "${instruction}" appended to "${name}".`;
        },
      }),
  });
