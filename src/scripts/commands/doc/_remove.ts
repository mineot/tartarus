import db from '../../../db';
import { CommandDoc } from '../../../types';
import { Feedback } from '../../../utils/feedback';

/**
 * Removes a description from a command.
 * @param {string} commandName - The name of the command to remove the description from.
 */
export async function removeDocCommand(commandName: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(commandName)) as CommandDoc;

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
