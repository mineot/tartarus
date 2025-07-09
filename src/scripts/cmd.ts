import db from '../db';
import { Feedback } from '../utils/feedback';

/**
 * Creates a new command with the given name and instruction.
 * @param {string} name The name of the command to create.
 * @param {string} instruction The instruction to add to the command.
 */
async function createNewCommand(name: string, instruction: string) {
  try {
    const command = await db.get(name).catch(() => null);

    if (command) {
      // If the command already exists, do nothing and alert the user.
      Feedback.notAllowed('Command already exists');
      return;
    }

    // Create a new document with the given name and instruction.
    await db.put({ _id: name, instructions: [instruction] });
    Feedback.success(`Command "${name}" created successfully`);
  } catch (error: any) {
    // If there's an error, log it and alert the user.
    Feedback.error(`Failed to create command: ${error.message}`);
  }
}

/**
 * Removes a command by its name from the database.
 * @param {string} name The name of the command to remove.
 */
async function removeCommand(name: string) {
  try {
    // Retrieve the command document by its name.
    const command = await db.get(name);
    // Remove the command document from the database.
    await db.remove(command);
    // Notify the user of successful removal.
    Feedback.success(`Command "${name}" removed successfully`);
  } catch (error: any) {
    // Handle the case where the command is not found.
    if (error.status === 404) {
      Feedback.notAllowed(`Command "${name}" not found`);
    } else {
      // Log an error message if the removal fails for another reason.
      Feedback.error(`Failed to remove command "${name}": ${error.message}`);
    }
  }
}

/**
 * Lists all commands registered in the database.
 * Prints a message with the number of registered commands if the database is empty.
 * Otherwise, prints the list of commands with their instructions.
 */
export default async function listCommands() {
  const { rows } = await db.allDocs({ include_docs: true });

  if (rows.length === 0) {
    Feedback.info('No commands registered.');
    return;
  }

  Feedback.message('Registered commands:');

  // Iterate over each command document in the result.
  for (const { doc } of rows) {
    if (doc) {
      // Print the name of the command.
      Feedback.message(`\nCommand: ${doc._id}`);

      // Iterate over each instruction in the command and print it.
      doc.instructions.forEach((instruction, index) => {
        Feedback.message(`- Instruction ${index} => ${instruction}`);
      });
    }
  }
}

/**
 * Registers the command group for the `cmd` command.
 * @param {import('commander').Command} program The Commander program.
 */
export function registerCmdGroup(program: import('commander').Command) {
  const command = program
    .command('cmd')
    .description('Manage named command sets: create, add instructions, or remove them.');

  /**
   * Creates a new command with the given name and instruction.
   * @param {string} name The name of the new command.
   * @param {string} instruction The instruction to add to the new command.
   */
  command
    .command('create')
    .argument('<name>', 'The name of the new command')
    .argument('<instruction>', 'The instruction to add to the new command')
    .description('Create a new command')
    .action(createNewCommand);

  /**
   * Removes a command by its name from the database.
   * @param {string} name The name of the command to remove.
   */
  command
    .command('remove')
    .argument('<name>', 'The name of the command to remove')
    .description('Delete a command')
    .action(removeCommand);

  /**
   * Lists all commands registered in the database.
   */
  command.command('list').description('List all commands').action(listCommands);
}
