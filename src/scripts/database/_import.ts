import { Command } from 'commander';
import db from 'src/core/db';
import fs from 'fs';
import path from 'path';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerDbImport = (program: Command) =>
  register({
    program,
    commandName: 'import',
    commandDescription: 'Restoring database from JSON backup.',
    commandHelp: {
      structure: [
        {
          name: 'path',
          description: 'Provide the path to the JSON file.',
        },
      ],
      example: `tartarus db import backup.json`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'db import',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the path to the JSON file.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [filePath] = args;
          const resolvedPath = path.resolve(filePath);

          if (!fs.existsSync(resolvedPath)) {
            FailThrow(`File not found: ${resolvedPath}`);
            return null;
          }

          const fileContent = fs.readFileSync(resolvedPath, 'utf-8');
          const docs = JSON.parse(fileContent);

          if (!Array.isArray(docs)) {
            FailThrow(`Invalid backup format. Expected an array of documents.`);
            return null;
          }

          const cleanDocs = docs.map((doc) => {
            const { _rev, ...rest } = doc;
            return rest;
          });

          await db.bulkDocs(cleanDocs);
          return `Backup imported from: ${resolvedPath}`;
        },
      }),
  });
