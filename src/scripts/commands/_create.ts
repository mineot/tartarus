import { Feedback } from '../../utils/feedback';
import db from '../../db';

/**
 * Creates a new command with the given name and instruction.
 * @param {string} name The name of the command to create.
 * @param {string} instruction The instruction to add to the command.
 */
export async function createCommand(name: string, instruction: string) {
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
