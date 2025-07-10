import { Command } from 'commander';
import { CommandDoc } from '../types';
import { Feedback } from '../utils/feedback';
import db from '../db';

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
    Feedback.success(`Command "${name}" with instruction "${instruction}" created successfully`);
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
      Feedback.warn(`Command "${name}" not found`);
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

  // TODO: add here the description if exists
  // Iterate over each command document in the result.
  for (const { doc } of rows) {
    if (doc) {
      // Print the name of the command.
      Feedback.message(`\nCommand: ${doc._id}`);

      const commandDoc = doc as CommandDoc;

      // Iterate over each instruction in the command and print it.
      commandDoc.instructions.forEach((instruction, index) => {
        Feedback.message(`- [${index}]: ${instruction}`);
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
    const commandDoc = (await db.get(name)) as CommandDoc;

    // Check if there are no instructions and notify the user.
    if (commandDoc.instructions.length === 0) {
      Feedback.warn('No instructions available.');
      return;
    }

    // Display the command name.
    Feedback.message(`Command: ${commandDoc._id}`);

    // Iterate over each instruction and display it.
    commandDoc.instructions.forEach((instruction, index) => {
      Feedback.message(`- [${index}]: ${instruction}`);
    });
  } catch {
    // Handle the case where the command is not found.
    Feedback.error(`Command "${name}" could not be found.`);
  }
}

export async function addInstructionCommand(name: string, newInstruction: string): Promise<void> {
  try {
    const commandDoc = (await db.get(name)) as CommandDoc;
    commandDoc.instructions.push(newInstruction);
    await db.put(commandDoc);
    Feedback.success(`Instruction "${newInstruction}" added to "${name}".`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.warn(`Command "${name}" not found`);
    } else {
      Feedback.error(`Failed to add instruction to "${name}": ${error.message}`);
    }
  }
}

/**
 * Removes an instruction from a command set.
 * @param {string} name The name of the command set to remove the instruction from.
 * @param {string} instructionIndex The index of the instruction to remove.
 * @returns {Promise<void>} A promise that resolves when the instruction has been removed.
 */
async function removeInstructionCommand(name: string, instructionIndex: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(name)) as CommandDoc;

    // Parse the instruction index and validate it.
    const index = parseInt(instructionIndex, 10);
    if (isNaN(index) || index < 0 || index >= commandDoc.instructions.length) {
      Feedback.error(`Invalid instruction index for "${name}".`);
      return;
    }

    // Remove the instruction at the specified index.
    const [removedInstruction] = commandDoc.instructions.splice(index, 1);

    // Save the updated command document to the database.
    await db.put(commandDoc);

    // Notify the user of the successful removal.
    Feedback.success(`Removed instruction "${removedInstruction}" from "${name}"`);
  } catch {
    // Handle the case where the command is not found.
    Feedback.error(`Failed to remove instruction from "${name}".`);
  }
}

/**
 * Registers the command group for the `cmd` command.
 * @param {import('commander').Command} program The Commander program.
 */
export function registerCmdGroup(program: Command): void {
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
    .action(createNewCommand);

  /**
   * Removes a command by its name from the database.
   * @param {string} name The name of the command to remove.
   */
  cmd
    .command('remove')
    .argument('<name>', 'Name of the command to remove')
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
    .argument('<name>', 'Name of the command to show')
    .description('Show the instructions of a command')
    .action(showCommand);

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
    .action(addInstructionCommand);

  /**
   * Removes an instruction from a command by its name.
   * @param {string} name The name of the command to remove an instruction from.
   * @param {string} instructionIndex The index of the instruction to remove.
   */
  cmd
    .command('subtract')
    .argument('<name>', 'Name of the command')
    .argument('<instructionIndex>', 'Index of the instruction to be subtracted')
    .description('Subtract an instruction from a command')
    .action(removeInstructionCommand);
}
