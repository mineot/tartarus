import { COMMAND_PREFIX } from '../__constants';
import { CommandDoc } from '../../../types';
import { Feedback } from '../../../utils/feedback';
import db from '../../../db';

export async function addDocCommand(commandName: string, descriptionText: string): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${commandName}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    commandDoc.description = descriptionText;
    await db.put(commandDoc);

    Feedback.success(`Description added to "${commandName}".`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${commandName}" not found`);
    } else {
      Feedback.error(`Failed to add description to "${commandName}": ${error.message}`);
    }
  }
}
