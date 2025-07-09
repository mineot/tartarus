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
 * Displays the details of a command by its name.
 * If the command is found, it lists all its instructions.
 * If the command is not found, or if there are no instructions,
 * appropriate feedback messages are displayed.
 * @param {string} name - The name of the command to display.
 */
async function showCommand(name: string) {
  try {
    // Retrieve the command document by its name.
    const commandDoc = await db.get(name);

    // Check if there are no instructions and notify the user.
    if (commandDoc.instructions.length === 0) {
      Feedback.warn('No instructions available.');
      return;
    }

    // Display the command name.
    Feedback.message(`Command: ${commandDoc._id}\n`);

    // Iterate over each instruction and display it.
    commandDoc.instructions.forEach((instruction, index) => {
      Feedback.message(`- Instruction ${index}: ${instruction}`);
    });
  } catch {
    // Handle the case where the command is not found.
    Feedback.error(`Command "${name}" could not be found.`);
  }
}

/**
 * Registers the command group for the `cmd` command.
 * @param {import('commander').Command} program The Commander program.
 */
export function registerCmdGroup(program: import('commander').Command): void {
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
    .argument('<name>', 'The name of the new command')
    .argument('<instruction>', 'The instruction to add to the new command')
    .description('Create a new command')
    .action(createNewCommand);

  /**
   * Removes a command by its name from the database.
   * @param {string} name The name of the command to remove.
   */
  cmd
    .command('remove')
    .argument('<name>', 'The name of the command to remove')
    .description('Delete a command')
    .action(removeCommand);

  /**
   * Lists all commands registered in the database.
   */
  cmd.command('list').description('List all commands').action(listCommands);

  /**
   * Shows the instructions of a command by its name.
   * @param {string} name The name of the command to show.
   */
  cmd
    .command('show')
    .argument('<name>', 'The name of the command to show')
    .description('Show the instructions of a command')
    .action(showCommand);
}
