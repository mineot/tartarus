import fs from 'fs';
import path from 'path';
import db from '../../db';
import { Feedback } from '../../utils/feedback';

/**
 * Imports commands from a JSON file into the database.
 * The JSON file should contain an array of CommandDoc objects.
 *
 * @param {string} filePath The path to the JSON file containing the commands.
 * @returns {Promise<void>} A promise that resolves when the import is completed.
 */
export async function importCommand(filePath: string): Promise<void> {
  // Resolve the absolute path of the file.
  const resolvedFilePath = path.resolve(filePath);

  // Check if the file exists at the specified path.
  if (!fs.existsSync(resolvedFilePath)) {
    Feedback.warn(`File not found: ${resolvedFilePath}`);
    return;
  }

  try {
    // Read the content of the file as a UTF-8 encoded string.
    const fileContent = fs.readFileSync(resolvedFilePath, { encoding: 'utf-8' });

    // Check if the file content is a valid JSON string.
    if (typeof fileContent !== 'string') {
      Feedback.error(`The file at ${resolvedFilePath} does not contain a valid JSON string.`);
      return;
    }

    // Parse the JSON string to an array of CommandDoc objects.
    const commands = JSON.parse(fileContent);

    // Check if the parsed value is an array of CommandDoc objects.
    if (!Array.isArray(commands)) {
      Feedback.error(
        `The file at ${resolvedFilePath} does not contain a valid JSON array of CommandDoc objects.`
      );
      return;
    }

    // Clean the commands by removing the _rev property.
    const cleaned = commands.map((cmd) => {
      const { _rev, ...rest } = cmd;
      return rest;
    });

    // Insert the parsed command documents into the database in bulk.
    await db.bulkDocs(cleaned);

    // Provide feedback of successful import.
    Feedback.success(`Backup imported from: ${resolvedFilePath}`);
  } catch (error: any) {
    // Log an error message if the import operation fails.
    Feedback.error(`Error importing the database: ${error.message}`);
  }
}
