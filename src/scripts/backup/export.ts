import { CommandDoc } from '../../types';
import db from '../../db';
import fs from 'fs';
import path from 'path';

/**
 * Exports all commands to a JSON file.
 *
 * @param {string} filePath The path to the JSON file to save the export to.
 * @returns {Promise<void>} A promise that resolves when the export is completed.
 */
export async function exportCommand(filePath: string): Promise<void> {
  try {
    const result = await db.allDocs({ include_docs: true });

    if (!result.rows.length) {
      console.log('No commands to export.');
      return;
    }

    const commands = result.rows.reduce((acc, row) => {
      // The `doc` property is only present if the document exists.
      if (row.doc) acc.push(row.doc);
      return acc;
    }, [] as CommandDoc[]);

    const resolvedPath = path.resolve(filePath);
    // Create the directory if it doesn't exist.
    fs.mkdirSync(path.dirname(resolvedPath), { recursive: true });
    // Write the JSON file.
    fs.writeFileSync(resolvedPath, JSON.stringify(commands, null, 2));

    console.log(`✅ Backup exported to: ${resolvedPath}`);
  } catch (error) {
    // Print the error message if something goes wrong.
    console.error('❌ Failed to export commands:', (error as Error).message);
  }
}
