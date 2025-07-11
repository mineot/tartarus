import { COMMAND_PREFIX } from './__constants';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';
import db from '../../db';

export async function listCommand() {
  Feedback.title('List All Registered Commands:');

  const rows = (await db.allDocs({ include_docs: true })).rows.filter((row) =>
    row.id.startsWith(COMMAND_PREFIX)
  );

  if (rows.length === 0) {
    Feedback.notFound('No commands registered.');
    return;
  }

  for (const { doc } of rows) {
    const commandDoc = doc as CommandDoc;

    const cmdId = commandDoc._id.replace(COMMAND_PREFIX, '');

    if (commandDoc.description) {
      Feedback.text(cmdId);
      Feedback.enphased(commandDoc.description);
    } else {
      Feedback.text(cmdId);
    }

    Feedback.break();

    commandDoc.instructions.forEach((instruction, index) => {
      Feedback.item(`[${index}]: ${instruction}`);
    });

    Feedback.break();
  }
}
