import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import { output } from 'src/utils/outputs';
import db from 'src/core/db';

export const registerCmdList = (program: Command) =>
  register({
    program,
    commandName: 'list',
    commandDescription: 'List all registered commands and their instructions',
    commandArguments: [],
    commandInstance: async () => {
      try {
        const rows = (await db.allDocs({ include_docs: true })).rows.filter((row) =>
          row.id.startsWith(COMMAND_PREFIX)
        );

        if (rows.length === 0) {
          FailThrow('No commands registered.');
          return;
        }

        for (const { doc } of rows) {
          const commandDoc = doc as CommandDoc;
          const id = commandDoc._id.replace(COMMAND_PREFIX, '');

          output({
            title: id,
            description: commandDoc.description ?? 'Below is the list of instructions',
            list: commandDoc.instructions.map((instruction, index) => {
              return `[${index}]: ${instruction}`;
            }),
          });
        }
      } catch (err: any) {
        error(err);
      }
    },
  });
