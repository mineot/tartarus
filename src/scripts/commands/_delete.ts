import db from '../../db';
import { Feedback } from '../../utils/feedback';

/**
 * Removes a command by its name from the database.
 * @param {string} name The name of the command to remove.
 */
export async function deleteCommand(name: string) {
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
