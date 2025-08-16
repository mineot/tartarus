import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { execInstruction } from 'src/utils/shell';
import { output } from 'src/utils/outputs';
import { register } from 'src/utils/register';
import db from 'src/core/db';
import inquirer from 'inquirer';

export const registerRun = (program: Command) =>
  register({
    program,
    commandName: 'run',
    commandDescription: 'Run an existing command',
    commandArguments: [
      {
        name: 'name',
        description: 'The name of the command.',
      },
    ],
    commandInstance: async (args: any) => {
      try {
        await execution(args.name);
      } catch (err: any) {
        error(err);
      }
    },
  });

async function execution(name: string | undefined): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    for (const [index, instruction] of commandDoc.instructions.entries()) {
      output({ title: instruction, list: [] });
      await execInstruction(instruction);
    }
  } catch (error: any) {
    if (error.status === 404) {
      return await buildQuestions(await findByName(name));
    } else {
      throw error;
    }
  }
}

async function buildQuestions(names: string[]): Promise<void> {
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
