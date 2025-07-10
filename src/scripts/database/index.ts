import { clearCommand } from './_clear';
import { Command } from 'commander';
import { exportCommand } from './_export';
import { importCommand } from './_import';

/**
 * Registers the `database` command group.
 * @param {import('commander').Command} program - The Commander program to register the command group with.
 */
export function registerDatabase(program: Command): void {
  // Define a new command group for database management.
  const database = program.command('db').description('Database management commands');

  /**
   * Clears all commands from the database.
   * This command fetches all documents, marks them for deletion, and performs a bulk delete operation.
   * Provides feedback on the number of commands deleted or if the database was already empty.
   */
  database
    .command('clear')
    .description('Clear all commands from the database')
    .action(clearCommand);

  /**
   * Exports all commands to a specified JSON file.
   * The file path must be provided as an argument.
   */
  database
    .command('export')
    .argument('<path>', 'Path to save exported JSON file')
    .description('Export all commands to a JSON file')
    .action(exportCommand);

  /**
   * Imports commands from a JSON file.
   * The file path must be provided as an argument.
   */
  database
    .command('import')
    .argument('<path>', 'Path to the JSON file to import')
    .description('Import commands from a JSON file')
    .action(importCommand);
}
