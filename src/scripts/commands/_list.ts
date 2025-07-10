import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';
import db from '../../db';

/**
 * Lists all commands registered in the database.
 * Provides a message if no commands are found.
 * Otherwise, it displays each command, its description if available,
 * and all associated instructions.
 */
export async function listCommand() {
  // Retrieve all documents from the database, including their content.
  const { rows } = await db.allDocs({ include_docs: true });

  // Check if there are no commands registered.
  if (rows.length === 0) {
    Feedback.info('No commands registered.');
    return;
  }

  // Inform the user that registered commands will be displayed.
  Feedback.message('Registered commands:');

  // Iterate over each command document in the result.
  for (const { doc } of rows) {
    if (doc) {
      const commandDoc = doc as CommandDoc;

      // Print the name of the command.
      Feedback.message(`\nCommand: ${commandDoc._id}\n`);

      // If a description is available, display it.
      if (commandDoc.description) {
        Feedback.message(`Description: ${commandDoc.description}\n`);
      }

      // Iterate over each instruction in the command and print it.
      commandDoc.instructions.forEach((instruction, index) => {
        Feedback.message(`- [${index}]: ${instruction}`);
      });
    }
  }
}
