import db from "../../db";
import { CommandDoc } from "../../types";
import { Feedback } from "../../utils/feedback";

/**
 * Adds a new instruction to the command set with the given name.
 * @param {string} name The name of the command set to add the instruction to.
 * @param {string} newInstruction The instruction to add to the command set.
 * @returns {Promise<void>} A promise that resolves when the instruction has been added.
 */
export async function appendCommand(name: string, newInstruction: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(name)) as CommandDoc;

    // Add the new instruction to the command document's instructions array.
    commandDoc.instructions.push(newInstruction);

    // Save the updated command document back to the database.
    await db.put(commandDoc);

    // Notify the user of the successful addition of the instruction.
    Feedback.success(`Instruction "${newInstruction}" added to "${name}".`);
  } catch (error: any) {
    // Handle the case where the command is not found.
    if (error.status === 404) {
      Feedback.warn(`Command "${name}" not found`);
    } else {
      // Log an error message if the addition fails for another reason.
      Feedback.error(`Failed to add instruction to "${name}": ${error.message}`);
    }
  }
}
