import { Command } from 'commander';
import { Feedback } from '../utils/feedback';
import db from '../db';

/**
 * Adds a description to a command.
 * @param {string} commandName - The name of the command to add a description to.
 * @param {string} descriptionText - The description text to add.
 */
async function addDocCommand(commandName: string, descriptionText: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = await db.get(commandName);

    // Add the description to the command document.
    commandDoc.description = descriptionText;

    // Save the updated command document back to the database.
    await db.put(commandDoc);

    // Provide feedback that the description was successfully added.
    Feedback.success(`Description added to "${commandName}".`);
  } catch (error: any) {
    // If there's an error, log it and alert the user.
    Feedback.error(`Failed to add description to "${commandName}": ${error.message}`);
  }
}

/**
 * Removes a description from a command.
 * @param {string} commandName - The name of the command to remove the description from.
 */
async function removeDocCommand(commandName: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = await db.get(commandName);

    // Check if the command document has a description.
    // If it does not have a description, print a message and do nothing else.
    if (!commandDoc.description) {
      Feedback.info(`No description to remove from "${commandName}".`);
      return;
    }

    // Remove the description from the command document.
    delete commandDoc.description;

    // Save the updated command document back to the database.
    await db.put(commandDoc);

    // Provide feedback that the description was successfully removed.
    Feedback.success(`Description removed from "${commandName}".`);
  } catch (error: any) {
    // If there's an error, log it and alert the user.
    Feedback.error(`Failed to remove description from "${commandName}": ${error.message}`);
  }
}

/**
 * Shows the description of a command by its name.
 * If the command is not found, it warns the user.
 * If the command is found, it displays the description.
 * @param {string} [name] - The name of the command to show the description for.
 */
async function showDocCommand(name?: string) {
  if (name) {
    try {
      // Retrieve the command document by its name.
      const doc = await db.get(name);

      // Check if the command document has a description.
      // If it does not have a description, print a message and do nothing else.
      if (!doc.description) {
        Feedback.warn(`Command "${name}" not found or has no description.`);
        return;
      }

      // Display the description of the command.
      Feedback.message(`"${name}" command documentation:`);
      Feedback.message(doc.description);
    } catch {
      Feedback.warn(`Command "${name}" not found.`);
    }
    return;
  }

  // Retrieve all documents from the database including their content.
  const allDocs = await db.allDocs({ include_docs: true });

  // Filter the documents to keep only those with a description.
  const described = allDocs.rows.map((row) => row.doc).filter((doc) => doc?.description);

  // Check if there are no documented commands.
  if (described.length === 0) {
    Feedback.info('No documented commands found.');
    return;
  }

  // Display the description of each documented command.
  for (const doc of described) {
    Feedback.message(`"${doc?._id}" command documentation:`);
    Feedback.message(`${doc?.description}\n`);
  }
}

/**
 * Registers the command group for managing command documentation.
 * @param {Command} program - The Commander program to register the command group with.
 */
export function registerDocGroup(program: Command): void {
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
