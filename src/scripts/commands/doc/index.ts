import { addDocCommand } from './_add';
import { Command } from 'commander';
import { removeDocCommand } from './_remove';
import { showDocCommand } from './_show';

export function registerCmdDoc(program: Command): void {
  const doc = program.command('doc').description('Manage command documentation');

  doc
    .command('add')
    .argument('<name>', 'Command name')
    .argument('<text>', 'Description text')
    .description('Add a description to a command')
    .action(addDocCommand);

  doc
    .command('remove')
    .argument('<name>', 'Command name')
    .description('Remove a description from a command')
    .action(removeDocCommand);

  doc
    .command('show')
    .argument('[name]', 'Command name (optional)')
    .description('Show command documentation')
    .action(showDocCommand);
}
