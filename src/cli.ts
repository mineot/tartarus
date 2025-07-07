import { Command } from 'commander';

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
program.command('export').action(exportCommand);
program.command('import').action(importCommand);
program.parse();
