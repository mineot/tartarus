import db from '../../db';

/**
 * Removes a command from the database.
 * Logs a success message if the command is found and removed.
 * Logs an error message if the command is not found.
 *
 * @param {string} name - The name of the command to remove.
 */
export async function removeCommand(name: string) {
  try {
    // Attempt to retrieve the command document by its name.
    const doc = await db.get(name);

    // Remove the command document from the database.
    await db.remove(doc);

    // Log a success message if the command was successfully removed.
    console.log(`✅ Command "${name}" removed.`);
  } catch {
    // Log an error message if the command is not found.
    console.error(`❌ Command "${name}" not found.`);
  }
}
