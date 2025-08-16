import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { done } from 'src/utils/outputs';
import { register } from 'src/utils/register';
import db from 'src/core/db';
import { error } from 'src/utils/error';

export const registerCmdAppend = (program: Command) =>
  register({
    program,
    commandName: 'append',
    commandDescription: 'Append a new instruction to a command',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the command.',
        required: true,
      },
      {
        name: 'instruction',
        description: 'Provide the new instruction to add to the command.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name, instruction } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const commandDoc = (await db.get(prefixName)) as CommandDoc;

        commandDoc.instructions.push(instruction);
        await db.put(commandDoc);

        done(`Instruction "${instruction}" appended to "${name}".`);
      } catch (err: any) {
        error(err);
      }
    },
  });
