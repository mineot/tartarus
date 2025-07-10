import db from '../../db';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';

/**
 * Displays the details of a command by its name.
 * If the command is found, it lists all its instructions.
 * If the command is not found, or if there are no instructions,
 * appropriate feedback messages are displayed.
 * @param {string} name - The name of the command to display.
 */
export async function showCommand(name: string) {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(name)) as CommandDoc;

    // Check if there are no instructions and notify the user.
    if (commandDoc.instructions.length === 0) {
      Feedback.warn('No instructions available.');
      return;
    }

    // Display the command name.
    Feedback.message(`Command: ${commandDoc._id}\n`);

    if (commandDoc.description) {
      // If a description is available, display it.
      Feedback.message(`Description: ${commandDoc.description}\n`);
    }

    // Iterate over each instruction and display it.
    commandDoc.instructions.forEach((instruction, index) => {
      Feedback.message(`- [${index}]: ${instruction}`);
    });
  } catch {
    // Handle the case where the command is not found.
    Feedback.error(`Command "${name}" could not be found.`);
  }
}
