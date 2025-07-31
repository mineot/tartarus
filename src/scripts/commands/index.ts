import { appendCommand } from './_append';
import { Command } from 'commander';
import { createCommand } from './_create';
import { deleteCommand } from './_delete';
import { listCommand } from './_list';
import { registerDocumentationCommandsToCommand } from './doc';
import { showCommand } from './_show';
import { subtractCommand } from './_subtract';

export function registerCMD(program: Command): void {
  const cmd = program
    .command('cmd')
    .description('Manage named command sets: create, add instructions, or remove them.');

  cmd
    .command('create')
    .argument('[name]', 'Name of the new command')
    .argument('[instruction]', 'Instruction to add to the new command')
    .description('Create a new command')
    .action(createCommand);

  cmd
    .command('append')
    .argument('[name]', 'Name of the command')
    .argument('[instruction]', 'Instruction to append to the command')
    .description('Append an instruction to a command')
    .action(appendCommand);

  cmd
    .command('subtract')
    .argument('[name]', 'Name of the command')
    .argument('[index]', 'Index of the instruction to remove')
    .description('Remove an instruction from a command')
    .action(subtractCommand);

  cmd
    .command('delete')
    .argument('[name]', 'Name of the command to remove')
    .description('Delete a command')
    .action(deleteCommand);

  cmd
    .command('show')
    .argument('[name]', 'Name of the command to show')
    .description('Show the instructions of a command')
    .action(showCommand);

  cmd.command('list').description('List all commands').action(listCommand);

  registerDocumentationCommandsToCommand(cmd);
}
