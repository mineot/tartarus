import { Command } from 'commander';
import { ConfigDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { MANUAL_EDITOR_CONFIG_ID } from 'src/utils/constants';
import { register } from 'src/utils/register';
import { output } from 'src/utils/outputs';
import db from 'src/core/db';

export const registerManShowEditor = (program: Command) =>
  register({
    program,
    commandName: 'show_editor',
    commandDescription: 'Show current manual editor',
    commandArguments: [],
    commandInstance: async (args: any) => {
      try {
        const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;
        if (config.data && typeof config.data === 'string') {
          output({
            title: 'Current editor',
            description: config.data,
            list: [],
          });
        } else {
          FailThrow('No editor configured. Use: tartarus man set_editor <editor>');
        }
      } catch (err: any) {
        error(err);
      }
    },
  });
