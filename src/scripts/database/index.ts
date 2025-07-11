import { clearCommand } from './_clear';
import { Command } from 'commander';
import { exportCommand } from './_export';
import { importCommand } from './_import';

export function registerDatabase(program: Command): void {
  const database = program.command('db').description('Database management commands');

  database
    .command('clear')
    .description('Clear all commands from the database')
    .action(clearCommand);

  database
    .command('export')
    .argument('<path>', 'Path to save exported JSON file')
    .description('Export all commands to a JSON file')
    .action(exportCommand);

  database
    .command('import')
    .argument('<path>', 'Path to the JSON file to import')
    .description('Import commands from a JSON file')
    .action(importCommand);
}
