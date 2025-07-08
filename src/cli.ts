import { Command } from 'commander';

import addCommand from './commands/add';
import clearCommand from './commands/clear';
import executeCommand from './commands/execute';
import exportCommand from './commands/export';
import importCommand from './commands/import';
import listCommand from './commands/list';
import registerCommand from './commands/register';
import removeCommand from './commands/remove';

const program = new Command();

program.name('tartarus').description('CLI for named commands').version('1.0.0');

program.command('register').argument('<name>').argument('<cmd>').action(registerCommand);
program.command('exec').argument('<name>').action(executeCommand);
program.command('list').action(listCommand);
program.command('remove').argument('<name>').action(removeCommand);
program.command('clear').description('Remove all commands from the database').action(clearCommand);

program
  .command('export')
  .argument('<path>', 'Path to save exported JSON file')
  .action(exportCommand);

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
