import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';
import { registerExecCommand } from './scripts/exec';

import { exportCommand } from './scripts/backup/export';
import addDocCommand from './scripts/docs/add';
import importCommand from './scripts/backup/import';
import remDocCommand from './scripts/docs/remove';
import showDocCommand from './scripts/docs/show';
import { registerDatabaseCommand } from './scripts/database';

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

// // BACKUP
// program
//   .command('export')
//   .argument('<path>', 'Path to save exported JSON file')
//   .description('Export all commands to a JSON file')
//   .action(exportCommand);

// program
//   .command('import')
//   .argument('<path>', 'Path to the JSON file to import')
//   .action(importCommand);

// program
//   .command('adddoc')
//   .argument('<name>', 'Command name')
//   .argument('<text>', 'Description text')
//   .action(addDocCommand);

// program.command('remdoc').argument('<name>', 'Command name').action(remDocCommand);

// program.command('showdoc').argument('[name]', 'Command name (optional)').action(showDocCommand);
