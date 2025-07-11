import { COMMAND_PREFIX } from '../__constants';
import { CommandDoc } from '../../../types';
import { Feedback } from '../../../utils/feedback';
import db from '../../../db';

export async function removeDocCommand(commandName: string): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${commandName}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    if (!commandDoc.description) {
      Feedback.warn(`No description to remove from "${commandName}".`);
      return;
    }

    delete commandDoc.description;

    await db.put(commandDoc);

    Feedback.success(`Description removed from "${commandName}".`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${commandName}" not found`);
    } else {
      Feedback.error(`Failed to remove description from "${commandName}": ${error.message}`);
    }
  }
}
