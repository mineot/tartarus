import { COMMAND_PREFIX } from 'src/scripts/commands/__constants';
import { CommandDoc } from 'src/types';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

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
