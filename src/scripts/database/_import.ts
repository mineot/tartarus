import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';
import fs from 'fs';
import path from 'path';

export const registerDbImport = (program: Command) =>
  register({
    program,
    commandName: 'import',
    commandDescription: 'Restoring database from JSON backup.',
    commandArguments: [
      {
        name: 'jsonFilePath',
        description: 'Provide the path to the JSON file.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { jsonFilePath } = args;
        const resolvedPath = path.resolve(jsonFilePath);

        if (!fs.existsSync(resolvedPath)) {
          FailThrow(`File not found: ${resolvedPath}`);
        }

        const fileContent = fs.readFileSync(resolvedPath, 'utf-8');
        const docs = JSON.parse(fileContent);

        if (!Array.isArray(docs)) {
          FailThrow(`Invalid backup format. Expected an array of documents.`);
        }

        const cleanDocs = docs.map((doc: any) => {
          const { _rev, ...rest } = doc;
          return rest;
        });

        await db.bulkDocs(cleanDocs);

        done(`Backup imported from: ${resolvedPath}`);
      } catch (err: any) {
        error(err);
      }
    },
  });
