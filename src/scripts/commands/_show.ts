import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import { output } from 'src/utils/outputs';
import db from 'src/core/db';

export const registerCmdShow = (program: Command) =>
  register({
    program,
    commandName: 'show',
    commandDescription: 'Show command details',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the command.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const commandDoc = (await db.get(prefixName)) as CommandDoc;

        if (commandDoc.instructions.length === 0) {
          FailThrow(`No instructions available for the command "${name}".`);
        }

        output({
          title: `Command ${name}`,
          description: commandDoc.description ?? 'Below is the list of instructions',
          list: commandDoc.instructions.map((instruction, index) => {
            return `[${index}]: ${instruction}`;
          }),
        });
      } catch (err: any) {
        error(err);
      }
    },
  });
