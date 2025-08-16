import { Command } from 'commander';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { register } from 'src/utils/register';
import db from 'src/core/db';
import fs from 'fs';
import path from 'path';

export const registerDbExport = (program: Command) =>
  register({
    program,
    commandName: 'export',
    commandDescription: 'Backup database to JSON file.',
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
        const result = await db.allDocs({ include_docs: true });
        const allDocs = result.rows.map((row) => row.doc).filter(Boolean);
        fs.writeFileSync(resolvedPath, JSON.stringify(allDocs, null, 2));
        done(`Backup exported to ${resolvedPath}`);
      } catch (err: any) {
        error(err);
      }
    },
  });
