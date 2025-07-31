import { COMMAND_PREFIX } from 'src/utils/constants';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

export async function createCommand(name: string, instruction: string) {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const command = await db.get(prefixName).catch(() => null);

    if (command) {
      Feedback.exists(`Command "${name}" already exists`);
      return;
    }

    Feedback.title(`Create Command: ${name}`);

    await db.put({ _id: prefixName, instructions: [instruction] });

    Feedback.success(`Command "${name}" with instruction "${instruction}" created successfully`);
  } catch (error: any) {
    Feedback.error(`Failed to create command: ${error.message}`);
  }
}
