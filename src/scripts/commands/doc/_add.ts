import db from '../../../db';
import { CommandDoc } from '../../../types';
import { Feedback } from '../../../utils/feedback';

/**
 * Adds a description to a command.
 * @param {string} commandName - The name of the command to add a description to.
 * @param {string} descriptionText - The description text to add.
 */
export async function addDocCommand(commandName: string, descriptionText: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(commandName)) as CommandDoc;

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
