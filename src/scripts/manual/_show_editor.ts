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
  TitledText,
  ValidationReturn,
} from 'src/utils/old-command';

export const registerManShowEditor = (program: Command) =>
  register({
    program,
    commandName: 'show_editor',
    commandDescription: 'Show current manual editor',
    commandHelp: {
      structure: [],
      example: `tartarus man show_editor`,
    },
    commandInstance: (args: Args) =>
      command(args, {
        referenceName: 'man show_editor',
        noArguments: true,
        validation: async (args: Args): ValidationReturn => {
          if (args.length > 0) {
            FailThrow('No arguments required.');
          }
        },
        operation: async (): OperationReturn => {
          const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;

          if (config.data && typeof config.data === 'string') {
            TitledText('Current editor', config.data);
          } else {
            FailThrow('No editor configured. Use: tartarus man set_editor <editor>');
          }

          return null;
        },
      }),
  });
