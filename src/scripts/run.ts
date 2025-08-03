import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/types';
import { execSync } from 'child_process';
import db from 'src/db';

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
    commandDescription: 'Run a command',
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

// import { execCommand } from './exec';
// import db from 'src/db';
// import { Feedback } from 'src/utils/feedback';
// import { COMMAND_PREFIX } from 'src/utils/constants';
// import { CommandDoc } from 'src/types';
// import inquirer from 'inquirer';

// export async function runCommand(inputName?: string) {
//   const { rows } = await db.allDocs({ include_docs: true });

//   // Coleta apenas comandos válidos
//   const commands: CommandDoc[] = rows
//     .filter((row) => row.id.startsWith(COMMAND_PREFIX))
//     .map((row) => row.doc as CommandDoc);

//   if (commands.length === 0) {
//     Feedback.warn('No commands registered.');
//     return;
//   }

//   let matches: CommandDoc[] = [];

//   if (inputName) {
//     // Tenta correspondência exata
//     const exactMatch = commands.find((cmd) => cmd._id === `${COMMAND_PREFIX}${inputName}`);
//     if (exactMatch) {
//       await execCommand(inputName);
//       return;
//     }

//     // Tenta correspondência por prefixo
//     matches = commands.filter((cmd) => cmd._id.replace(COMMAND_PREFIX, '').startsWith(inputName));

//     if (matches.length === 0) {
//       Feedback.notFound(`No command found starting with "${inputName}"`);
//       return;
//     }
//   } else {
//     matches = commands;
//   }

//   // Lista para escolha via Inquirer
//   const { selected } = await inquirer.prompt([
//     {
//       type: 'list',
//       name: 'selected',
//       message: inputName
//         ? `Found ${matches.length} matches for "${inputName}". Select one to run:`
//         : 'Select a command to run:',
//       choices: matches.map((cmd, index) => ({
//         name: `${index + 1}. ${cmd._id.replace(COMMAND_PREFIX, '')}${cmd.description ? ' - ' + cmd.description : ''}`,
//         value: cmd._id.replace(COMMAND_PREFIX, ''),
//       })),
//     },
//   ]);

//   await execCommand(selected);
// }
