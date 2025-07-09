import db from '../../db';

/**
 * Registers a command with the given name and instruction.
 * If the command already exists, it will be updated with the new instruction.
 * @param name The name of the command to register.
 * @param command The instruction to add to the command.
 */
export async function registerCommand(name: string, command: string) {
  try {
    // First, check if the command already exists. If it does, update it.
    // Otherwise, create a new document.
    const existing = await db.get(name).catch(() => null);
    if (existing) {
      // Update the existing document.
      existing.instructions.push(command);
      await db.put(existing);
    } else {
      // Create a new document.
      await db.put({ _id: name, instructions: [command] });
    }
    console.log(`✅ Command "${name}" registered with 1 instruction.`);
  } catch (err: any) {
    // If there was an error, print it to the console.
    console.error('❌ Registration error:', err.message);
  }
}
