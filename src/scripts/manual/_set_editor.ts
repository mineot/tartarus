import { Command } from 'commander';
import { ConfigDoc } from 'src/core/types';
import { MANUAL_EDITOR_CONFIG_ID } from 'src/utils/constants';
import db from 'src/core/db';

import {
  Args,
  command,
  FailThrow,
  OperationReturn,
  register,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerManSetEditor = (program: Command) =>
  register({
    program,
    commandName: 'set_editor',
    commandDescription: 'Set the manual editor',
    commandHelp: {
      structure: [
        {
          name: 'editor',
          description: 'Provide the name of the editor. Example: nano, vim or notepad',
        },
      ],
      example: `tartarus man set_editor <editor>`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man set_editor',
        validation: async (args: Args): ValidationReturn => {
          if (args.length != 1) {
            FailThrow('You must provide one argument: the editor name.');
          }
        },
        operation: async (args: Args): OperationReturn => {
          const [editor] = args;

          if (!editor || !editor.trim()) {
            FailThrow(
              'You must provide a valid editor name (e.g. nano, vim, notepad).\nSee the help for more information.'
            );

            return null;
          }

          const config: ConfigDoc = {
            _id: MANUAL_EDITOR_CONFIG_ID,
            data: editor.trim(),
            updatedAt: new Date().toISOString(),
          };

          try {
            const existing = await db.get(MANUAL_EDITOR_CONFIG_ID);
            await db.put({ ...existing, ...config });
            return `Editor updated to "${editor}".`;
          } catch {
            await db.put(config);
            return `Editor set to "${editor}".`;
          }
        },
      }),
  });
