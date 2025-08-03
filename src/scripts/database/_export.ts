import { Command } from 'commander';
import db from 'src/db';
import fs from 'fs';
import path from 'path';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/command';

export const registerDbExport = (program: Command) =>
  register({
    program,
    commandName: 'export',
    commandDescription: 'Backup database to JSON file.',
    commandHelp: {
      structure: [
        {
          name: 'path',
          description: 'Provide the path to the JSON file.',
        },
      ],
      example: `tartarus db export backup.json`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'db export',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the path to the JSON file.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [filePath] = args;
          const resolvedPath = path.resolve(filePath);
          const result = await db.allDocs({ include_docs: true });
          const allDocs = result.rows.map((row) => row.doc).filter(Boolean);
          fs.writeFileSync(resolvedPath, JSON.stringify(allDocs, null, 2));
          return `Backup exported to ${resolvedPath}`;
        },
      }),
  });
