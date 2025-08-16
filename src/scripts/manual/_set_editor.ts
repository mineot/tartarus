import { Command } from 'commander';
import { ConfigDoc } from 'src/core/types';
import { done } from 'src/utils/outputs';
import { error } from 'src/utils/error';
import { MANUAL_EDITOR_CONFIG_ID } from 'src/utils/constants';
import { register } from 'src/utils/register';
import db from 'src/core/db';

export const registerManSetEditor = (program: Command) =>
  register({
    program,
    commandName: 'set_editor',
    commandDescription: 'Set the manual editor application (e.g. nano, vim, notepad)',
    commandArguments: [
      {
        name: 'editor',
        description: 'Provide the name of the editor. Example: nano, vim or notepad',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { editor } = args;

        const config: ConfigDoc = {
          _id: MANUAL_EDITOR_CONFIG_ID,
          data: editor.trim(),
          updatedAt: new Date().toISOString(),
        };

        try {
          const existing = await db.get(MANUAL_EDITOR_CONFIG_ID);
          await db.put({ ...existing, ...config });
          done(`Editor updated to "${editor}".`);
        } catch {
          await db.put(config);
          done(`Editor set to "${editor}".`);
        }
      } catch (err: any) {
        error(err);
      }
    },
  });
