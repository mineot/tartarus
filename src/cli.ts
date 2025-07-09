import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';

import { addCommand } from './scripts/data/instructions/add';
import { exportCommand } from './scripts/backup/export';
import addDocCommand from './scripts/docs/add';
import clearCommand from './scripts/utils/clear';
import executeCommand from './scripts/execute';
import importCommand from './scripts/backup/import';
import remDocCommand from './scripts/docs/remove';
import showDocCommand from './scripts/docs/show';
import removeInstruction from './scripts/data/instructions/remove';

const program = new Command();

program
  .name('tartarus')
  .version('1.0.0')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCmdGroup(program);

program.parse();

// program
//   .command('add')
//   .argument('<name>', 'Command name')
//   .argument('<instruction>', 'Instruction to add')
//   .description('Add an instruction to a command')
//   .action(addCommand);

// // BACKUP
// program
//   .command('export')
//   .argument('<path>', 'Path to save exported JSON file')
//   .description('Export all commands to a JSON file')
//   .action(exportCommand);

// program.command('exec').argument('<name>').action(executeCommand);
// program.command('clear').description('Remove all commands from the database').action(clearCommand);

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

// program
//   .command('rminst')
//   .argument('<name>', 'Command name')
//   .argument('<index>', 'Index of the instruction to remove (starting from 0)')
//   .action(removeInstruction);
