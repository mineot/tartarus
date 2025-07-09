import { Command } from 'commander';

// CRUD
import { registerCommand } from './commands/data/register';
import { addCommand } from './commands/data/instructions/add';
import { removeCommand } from './commands/data/remove';

// BACKUP
import { exportCommand } from './commands/backup/export';

import addDocCommand from './commands/docs/add';
import clearCommand from './commands/utils/clear';
import executeCommand from './commands/execute';
import importCommand from './commands/backup/import';
import listCommand from './commands/data/list';
import remDocCommand from './commands/docs/remove';
import showDocCommand from './commands/docs/show';
import removeInstruction from './commands/data/instructions/remove';

const program = new Command();

program.name('tartarus').description('CLI for named commands').version('1.0.0');

// CRUD

program
  .command('reg')
  .argument('<name>')
  .argument('<cmd>')
  .description('Register a new command')
  .action(registerCommand);

program.command('del').argument('<name>').description('Delete a command').action(removeCommand);

program
  .command('add')
  .argument('<name>', 'Command name')
  .argument('<instruction>', 'Instruction to add')
  .description('Add an instruction to a command')
  .action(addCommand);

// BACKUP
program
  .command('export')
  .argument('<path>', 'Path to save exported JSON file')
  .description('Export all commands to a JSON file')
  .action(exportCommand);

program.command('exec').argument('<name>').action(executeCommand);
program.command('list').action(listCommand);
program.command('clear').description('Remove all commands from the database').action(clearCommand);

program
  .command('import')
  .argument('<path>', 'Path to the JSON file to import')
  .action(importCommand);
program.parse();

program
  .command('adddoc')
  .argument('<name>', 'Command name')
  .argument('<text>', 'Description text')
  .action(addDocCommand);

program.command('remdoc').argument('<name>', 'Command name').action(remDocCommand);

program.command('showdoc').argument('[name]', 'Command name (optional)').action(showDocCommand);

program
  .command('rminst')
  .argument('<name>', 'Command name')
  .argument('<index>', 'Index of the instruction to remove (starting from 0)')
  .action(removeInstruction);
