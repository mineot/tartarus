import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { execInstruction } from 'src/utils/shell';
import db from 'src/core/db';
import inquirer from 'inquirer';

import {
  Args,
  command,
  FailThrow,
  ItemText,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/old-command';

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
        noArguments: true,
        validation: async (): ValidationReturn => {},
        operation: async (args: Args): OperationReturn => {
          const [name] = args;
          return execution(name);
        },
      }),
  });

async function execution(name: string | undefined): Promise<OperationReturn> {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    for (const [index, instruction] of commandDoc.instructions.entries()) {
      ItemText(index, instruction);
      await execInstruction(instruction);
    }

    return null;
  } catch (error: any) {
    if (error.status === 404) {
      return await buildQuestions(await findByName(name));
    } else {
      throw error;
    }
  }
}

async function buildQuestions(names: string[]): Promise<OperationReturn> {
  const { choice } = await inquirer.prompt([
    {
      type: 'list',
      name: 'choice',
      message: 'Choose a command to run:',
      choices: names,
    },
  ]);

  if (choice) {
    return execution(choice);
  }

  return null;
}

async function findByName(name: string | undefined): Promise<string[]> {
  const rows = (await db.allDocs({ include_docs: true })).rows.filter((row) =>
    row.id.startsWith(COMMAND_PREFIX)
  );

  const items = rows
    .filter((item) => (name === undefined ? true : item.doc?._id?.includes(name)))
    .map((item) => item.doc?._id.replace(COMMAND_PREFIX, ''))
    .filter((id): id is string => typeof id === 'string');

  if (items.length === 0) {
    FailThrow('No commands found.');
    return [];
  }

  return items;
}
