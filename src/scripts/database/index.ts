import { Command } from 'commander';
import { registerDbClear } from './_clear';
import { registerDbExport } from './_export';
import { registerDbImport } from './_import';

export function registerDatabase(program: Command): void {
  const database = program.command('db').description('Database management.');
  registerDbClear(database);
  registerDbExport(database);
  registerDbImport(database);
}
