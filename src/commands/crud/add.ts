import db from '../../db';

/**
 * Adds a new instruction to an existing command.
 * If the command does not exist, an error message is logged.
 * @param name The name of the command to update.
 * @param newInstruction The new instruction to add to the command.
 */
export async function addCommand(name: string, newInstruction: string) {
  try {
    // Retrieve the command document by its name.
    const doc = await db.get(name);
    // Add the new instruction to the command's instructions array.
    doc.instructions.push(newInstruction);
    // Update the command document in the database.
    await db.put(doc);
    console.log(`✅ Instruction added to "${name}".`);
  } catch {
    // Log an error message if the command is not found.
    console.error(`❌ Command "${name}" not found.`);
  }
}
