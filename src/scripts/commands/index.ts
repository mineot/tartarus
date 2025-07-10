import { appendCommand } from './_append';
import { Command } from 'commander';
import { createCommand } from './_create';
import { deleteCommand } from './_delete';
import { listCommand } from './_list';
import { registerCmdDoc } from './doc';
import { showCommand } from './_show';
import { subtractCommand } from './_subtract';

/**
 * Registers the command group for the `cmd` command.
 * @param {import('commander').Command} program The Commander program.
 */
export function registerCMD(program: Command): void {
  const cmd = program
    .command('cmd')
    .description('Manage named command sets: create, add instructions, or remove them.');

  /**
   * Creates a new command with the given name and instruction.
   * @param {string} name The name of the new command.
   * @param {string} instruction The instruction to add to the new command.
   */
  cmd
    .command('create')
    .argument('<name>', 'Name of the new command')
    .argument('<instruction>', 'Instruction to add to the new command')
    .description('Create a new command')
    .action(createCommand);

  /**
   * Adds an instruction to a command by its name.
   * @param {string} name The name of the command to add an instruction to.
   * @param {string} instruction The instruction to add to the command.
   */
  cmd
    .command('append')
    .argument('<name>', 'Name of the command')
    .argument('<instruction>', 'Instruction to append to the command')
    .description('Append an instruction to a command')
    .action(appendCommand);

  /**
   * Removes an instruction from a command by its name.
   * @param {string} name The name of the command to remove an instruction from.
   * @param {string} instructionIndex The index of the instruction to remove.
   */
  cmd
    .command('subtract')
    .argument('<name>', 'Name of the command')
    .argument('<instructionIndex>', 'Index of the instruction to remove')
    .description('Remove an instruction from a command')
    .action(subtractCommand);

  /**
   * Removes a command by its name from the database.
   * @param {string} name The name of the command to remove.
   */
  cmd
    .command('delete')
    .argument('<name>', 'Name of the command to remove')
    .description('Delete a command')
    .action(deleteCommand);

  /**
   * Shows the instructions of a command by its name.
   * @param {string} name The name of the command to show.
   */
  cmd
    .command('show')
    .argument('<name>', 'Name of the command to show')
    .description('Show the instructions of a command')
    .action(showCommand);

  /**
   * Lists all commands registered in the database.
   */
  cmd.command('list').description('List all commands').action(listCommand);

  /**
   * Registers the command group for managing command documentation.
   */
  registerCmdDoc(cmd);
}
