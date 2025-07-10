import fs from 'fs';
import path from 'path';
import db from '../../db';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';

/**
 * Exports the entire database to a JSON file.
 * @param {string} filePath The path where the JSON file should be saved.
 * @returns {Promise<void>} A promise that resolves when the export is completed.
 */
export async function exportCommand(filePath: string): Promise<void> {
  try {
    // Retrieve all documents from the database, including their content.
    const result = await db.allDocs({ include_docs: true });

    // Check if there are no commands to export.
    if (!result.rows.length) {
      Feedback.info('No commands to export.');
      return;
    }

    // Extract the command documents from the result and store them in an array.
    const commands = result.rows.reduce((acc, row) => {
      if (row.doc) acc.push(row.doc as CommandDoc);
      return acc;
    }, [] as CommandDoc[]);

    // Create the directory if it doesn't already exist.
    const resolvedPath = path.resolve(filePath);
    fs.mkdirSync(path.dirname(resolvedPath), { recursive: true });

    // Write the commands to the file as a JSON string with indentation.
    fs.writeFileSync(resolvedPath, JSON.stringify(commands, null, 2));

    Feedback.success(`Backup exported to: ${resolvedPath}`);
  } catch (error: any) {
    // Log an error message if the operation fails.
    Feedback.error(`Error exporting the database: ${error.message}`);
  }
}
