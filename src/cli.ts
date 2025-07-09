import { Command } from 'commander';

// BACKUP
import { exportCommand } from './commands/backup/export';

import addCommand from './commands/crud/add';
import addDocCommand from './commands/docs/add';
import clearCommand from './commands/utils/clear';
import executeCommand from './commands/execute';
import importCommand from './commands/backup/import';
import listCommand from './commands/crud/list';
import registerCommand from './commands/crud/register';
import remDocCommand from './commands/docs/remove';
import removeCommand from './commands/crud/remove';
import showDocCommand from './commands/docs/show';

const program = new Command();

program.name('tartarus').description('CLI for named commands').version('1.0.0');

// BACKUP
program
  .command('export')
  .argument('<path>', 'Path to save exported JSON file')
  .description('Export all commands to a JSON file')
  .action(exportCommand);

program.command('register').argument('<name>').argument('<cmd>').action(registerCommand);
program.command('exec').argument('<name>').action(executeCommand);
program.command('list').action(listCommand);
program.command('remove').argument('<name>').action(removeCommand);
program.command('clear').description('Remove all commands from the database').action(clearCommand);

program
  .command('import')
  .argument('<path>', 'Path to the JSON file to import')
  .action(importCommand);
program.parse();

program
  .command('add')
  .argument('<name>', 'Command name')
  .argument('<instruction>', 'Instruction to add')
  .action(addCommand);

program
  .command('adddoc')
  .argument('<name>', 'Command name')
  .argument('<text>', 'Description text')
  .action(addDocCommand);

program.command('remdoc').argument('<name>', 'Command name').action(remDocCommand);

program.command('showdoc').argument('[name]', 'Command name (optional)').action(showDocCommand);
