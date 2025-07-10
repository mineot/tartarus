import { Command } from 'commander';
import { addDocCommand } from './_add';
import { removeDocCommand } from './_remove';
import { showDocCommand } from './_show';

/**
 * Registers the command group for managing command documentation.
 * @param {Command} program - The Commander program to register the command group with.
 */
export function registerCmdDoc(program: Command): void {
  // Define a new command group for documentation management.
  const doc = program.command('doc').description('Manage command documentation');

  /**
   * Adds a description to a command.
   * Takes the command name and the description text as arguments.
   */
  doc
    .command('add')
    .argument('<name>', 'Command name')
    .argument('<text>', 'Description text')
    .description('Add a description to a command')
    .action(addDocCommand);

  /**
   * Removes a description from a command by its name.
   * Takes the command name as an argument.
   */
  doc
    .command('remove')
    .argument('<name>', 'Command name')
    .description('Remove a description from a command')
    .action(removeDocCommand);

  /**
   * Shows the documentation for a specific command or all commands if no name is provided.
   * Takes an optional command name as an argument.
   */
  doc
    .command('show')
    .argument('[name]', 'Command name (optional)')
    .description('Show command documentation')
    .action(showDocCommand);
}
