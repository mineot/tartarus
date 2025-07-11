import { COMMAND_PREFIX } from './__constants';
import { CommandDoc } from '../../types';
import { Feedback } from '../../utils/feedback';
import db from '../../db';

export async function subtractCommand(name: string, instructionIndex: string): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    const index = parseInt(instructionIndex, 10);

    if (isNaN(index) || index < 0 || index >= commandDoc.instructions.length) {
      Feedback.warn(`Invalid instruction index for "${name}".`);
      return;
    }

    const [removedInstruction] = commandDoc.instructions.splice(index, 1);

    await db.put(commandDoc);

    Feedback.success(`Removed instruction "${removedInstruction}" from "${name}"`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${name}" not found`);
    } else {
      Feedback.error(`Failed to add instruction to "${name}": ${error.message}`);
    }
  }
}
