import { Command } from 'commander';
import { CommandDoc } from '../types';
import { Feedback } from '../utils/feedback';
import db from '../db';
import fs from 'fs';
import path from 'path';

/**
 * Clears all commands from the database.
 * Fetches all documents, marks them for deletion, and then performs a bulk delete operation.
 * Provides feedback on the number of commands deleted or if the database was already empty.
 */
async function clearCommand() {
  try {
    // Retrieve all documents from the database including their content.
    const allDocs = await db.allDocs({ include_docs: true });

    // Filter and map the documents to mark them for deletion.
    const deletableDocs = allDocs.rows
      .map((row) => row.doc)
      .filter(
        (
          doc
        ): doc is PouchDB.Core.IdMeta &
          PouchDB.Core.RevisionIdMeta &
          CommandDoc & { _deleted?: boolean } => {
          // Ensure the document has valid _id and _rev properties.
          return doc !== undefined && typeof doc._id === 'string' && typeof doc._rev === 'string';
        }
      )
      .map((doc) => ({ ...doc, _deleted: true }));

    // Check if there are no documents to delete.
    if (deletableDocs.length === 0) {
      Feedback.info('No commands to delete. The database is empty.');
      return;
    }

    // Perform a bulk delete operation on the database.
    await db.bulkDocs(deletableDocs);
    Feedback.success(`Successfully deleted ${deletableDocs.length} command(s) from the database.`);
  } catch (error: any) {
    // Log an error message if the operation fails.
    Feedback.error(`Error clearing the database: ${error.message}`);
  }
}

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
      if (row.doc) acc.push(row.doc);
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

/**
 * Registers the `database` command group.
 * @param {import('commander').Command} program - The Commander program to register the command group with.
 */
export function registerDatabaseGroup(program: Command): void {
  // Define a new command group for database management.
  const database = program.command('db').description('Database management commands');

  /**
   * Clears all commands from the database.
   * This command fetches all documents, marks them for deletion, and performs a bulk delete operation.
   * Provides feedback on the number of commands deleted or if the database was already empty.
   */
  database
    .command('clear')
    .description('Clear all commands from the database')
    .action(clearCommand);

  /**
   * Exports all commands to a specified JSON file.
   * The file path must be provided as an argument.
   */
  database
    .command('export')
    .argument('<path>', 'Path to save exported JSON file')
    .description('Export all commands to a JSON file')
    .action(exportCommand);

  /**
   * Imports commands from a JSON file.
   * The file path must be provided as an argument.
   */
  database
    .command('import')
    .argument('<path>', 'Path to the JSON file to import')
    .description('Import commands from a JSON file')
    .action(importCommand);
}
