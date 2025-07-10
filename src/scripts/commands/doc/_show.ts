import db from '../../../db';
import { CommandDoc } from '../../../types';
import { Feedback } from '../../../utils/feedback';

/**
 * Shows the description of a command by its name.
 * If the command is not found, it warns the user.
 * If the command is found, it displays the description.
 * @param {string} [name] - The name of the command to show the description for.
 */
export async function showDocCommand(name?: string) {
  if (name) {
    try {
      // Retrieve the command document by its name.
      const doc = (await db.get(name)) as CommandDoc;

      // Check if the command document has a description.
      // If it does not have a description, print a message and do nothing else.
      if (!doc.description) {
        Feedback.warn(`Command "${name}" not found or has no description.`);
        return;
      }

      // Display the description of the command.
      Feedback.message(`"${name}" command documentation:`);
      Feedback.message(doc.description);
    } catch {
      Feedback.warn(`Command "${name}" not found.`);
    }
    return;
  }

  // Retrieve all documents from the database including their content.
  const allDocs = await db.allDocs({ include_docs: true });

  // Filter the documents to keep only those with a description.
  const described = allDocs.rows
    .map((row) => row.doc)
    .filter((doc) => (doc as CommandDoc)?.description);

  // Check if there are no documented commands.
  if (described.length === 0) {
    Feedback.info('No documented commands found.');
    return;
  }

  // Display the description of each documented command.
  for (const doc of described) {
    Feedback.message(`"${doc?._id}" command documentation:`);
    Feedback.message(`${(doc as CommandDoc)?.description}\n`);
  }
}
