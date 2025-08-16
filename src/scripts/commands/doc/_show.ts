import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/utils/constants';
import { CommandDoc } from 'src/core/types';
import { error, FailThrow } from 'src/utils/error';
import { register } from 'src/utils/register';
import { output } from 'src/utils/outputs';
import db from 'src/core/db';

export const registerShowDocumentation = (program: Command) =>
  register({
    program,
    commandName: 'show',
    commandDescription: 'Show command documentation',
    commandArguments: [
      {
        name: 'name',
        description: 'The name of the command to show documentation for.',
        required: true,
      },
    ],
    commandInstance: async (args: any) => {
      try {
        const { name } = args;
        const prefixName = `${COMMAND_PREFIX}${name}`;
        const doc = (await db.get(prefixName)) as CommandDoc;

        if (!doc.description) {
          FailThrow('No documentation to show.');
          return;
        }

        output({ title: name, list: [doc.description] });
      } catch (err: any) {
        error(err);
      }
    },
  });
