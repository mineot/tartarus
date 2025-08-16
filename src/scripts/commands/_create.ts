import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { done } from 'src/utils/outputs';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerCmdCreate = (program: Command) =>
  register({
    program,
    commandName: 'create',
    commandDescription: 'Create a new command',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the command.',
        required: true,
      },
      {
        name: 'instruction',
        description: 'Provide the first instruction to add to the command.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name, instruction } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const command = await db.get(prefixName).catch(() => null);
  
        if (command) {
          FailThrow(`Command "${name}" already exists`);
        }
  
        await db.put({ _id: prefixName, instructions: [instruction] });
  
        done(`Command "${name}" with instruction "${instruction}" was created.`);
      } catch (err: any) {
        error(err);
      }
    },
  });
