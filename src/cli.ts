import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';
import { registerExecCommand } from './scripts/exec';
import { registerDatabaseCommand } from './scripts/database';

import addDocCommand from './scripts/docs/add';
import remDocCommand from './scripts/docs/remove';
import showDocCommand from './scripts/docs/show';

const program = new Command();

program
  .name('tartarus')
  .version('1.0.0')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCmdGroup(program);
registerExecCommand(program);
registerDatabaseCommand(program);

program.parse();

// program
//   .command('adddoc')
//   .argument('<name>', 'Command name')
//   .argument('<text>', 'Description text')
//   .action(addDocCommand);

// program.command('remdoc').argument('<name>', 'Command name').action(remDocCommand);

// program.command('showdoc').argument('[name]', 'Command name (optional)').action(showDocCommand);
