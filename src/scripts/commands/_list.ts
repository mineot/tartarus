import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/types';
import db from 'src/db';

import {
  Args,
  BreakLine,
  command,
  FailThrow,
  ItemText,
  OperationReturn,
  register,
  TitledText,
  ValidationReturn,
} from 'src/utils/command';

export const registerCmdList = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List all registered commands and their instructions',
    commandHelp: {
      structure: [],
      example: `tartarus cmd list`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'cmd list',
        noArguments: true,
        validation: async (args: Args): ValidationReturn => {
          if (args.length > 0) {
            FailThrow('No arguments required.');
          }
        },
        operation: async (): OperationReturn => {
          const rows = (await db.allDocs({ include_docs: true })).rows.filter((row) =>
            row.id.startsWith(COMMAND_PREFIX)
          );

          if (rows.length === 0) {
            FailThrow('No commands registered.');
            return null;
          }

          for (const { doc } of rows) {
            const commandDoc = doc as CommandDoc;
            const id = commandDoc._id.replace(COMMAND_PREFIX, '');

            TitledText(
              `Command ${id}`,
              commandDoc.description ?? 'Below is the list of instructions'
            );

            commandDoc.instructions.forEach((instruction, index) => {
              ItemText(index, instruction);
            });

            BreakLine();
          }

          return null;
        },
      }),
  });
