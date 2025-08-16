import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { done } from 'src/utils/outputs';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerCmdSubtract = (program: Command) =>
  register({
    program,
    commandName: 'subtract',
    commandDescription: 'Subtract an instruction to a command',
    commandArguments: [
      {
        name: 'name',
        description: 'Provide the name of the command.',
        required: true,
      },
      {
        name: 'index',
        description: 'Provide the index of the instruction you want to remove.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name, index } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const commandDoc = (await db.get(prefixName)) as CommandDoc;
        const idx = parseInt(index, 10);

        if (isNaN(idx) || idx < 0 || idx >= commandDoc.instructions.length) {
          FailThrow(`Invalid instruction index for "${name}".`);
        }

        const [removedInstruction] = commandDoc.instructions.splice(idx, 1);
        await db.put(commandDoc);

        done(`Removed instruction "${removedInstruction}" from "${name}"`);
      } catch (err: any) {
        error(err);
      }
    },
  });

// import {
//   Args,
//   command,
//   FailThrow,
//   OperationReturn,
//   register,
//   ValidationReturn,
// } from 'src/utils/old-command';

// export const registerCmdSubtract = (program: Command) =>
//   register({
//     program,
//     commandName: 'subtract',
//     commandDescription: 'Subtract an instruction to a command',
//     commandHelp: {
//       structure: [
//         {
//           name: 'name',
//           description: 'Provide the name of the command.',
//         },
//         {
//           name: 'index',
//           description: 'Provide the index of the instruction you want to remove.',
//         },
//       ],
//       example: `tartarus cmd subtract commandName instructionIndex`,
//     },
//     commandInstance: (args: Args) =>
//       command(args, {
//         referenceName: 'cmd subtract',
//         validation: async (args: Args): ValidationReturn => {
//           if (args.length != 2) {
//             FailThrow(
//               'You must provide two arguments: the command name and the index of the instruction you want to remove.'
//             );
//           }
//         },
//         operation: async (args: Args): OperationReturn => {
//           const [name, index] = args;
//           const prefixName = `${COMMAND_PREFIX}${name}`;
//           const commandDoc = (await db.get(prefixName)) as CommandDoc;
//           const idx = parseInt(index, 10);

//           if (isNaN(idx) || idx < 0 || idx >= commandDoc.instructions.length) {
//             FailThrow(`Invalid instruction index for "${name}".`);
//             return null;
//           }

//           const [removedInstruction] = commandDoc.instructions.splice(idx, 1);
//           await db.put(commandDoc);

//           return `Removed instruction "${removedInstruction}" from "${name}"`;
//         },
//       }),
//   });
