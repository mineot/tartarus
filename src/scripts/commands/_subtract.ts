import db from '../../db';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';

/**
 * Removes an instruction from a command set.
 * @param {string} name The name of the command set to remove the instruction from.
 * @param {string} instructionIndex The index of the instruction to remove.
 * @returns {Promise<void>} A promise that resolves when the instruction has been removed.
 */
export async function subtractCommand(name: string, instructionIndex: string): Promise<void> {
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
