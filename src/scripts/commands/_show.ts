import { COMMAND_PREFIX } from './__constants';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';
import db from '../../db';

export async function showCommand(name: string) {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    if (commandDoc.instructions.length === 0) {
      Feedback.notFound(`No instructions available for the command "${name}".`);
      return;
    }

    Feedback.title(`Show Details for Command: ${name}`);

    if (commandDoc.description) {
      Feedback.enphased(commandDoc.description);
      Feedback.break();
    }

    commandDoc.instructions.forEach((instruction, index) => {
      Feedback.item(`[${index}]: ${instruction}`);
    });
  } catch (err: any) {
    if (err.status == 404) {
      Feedback.notFound(`Command "${name}" not found.`);
    } else {
      Feedback.error(err.message);
    }
  }
}
