import { COMMAND_PREFIX } from 'src/scripts/commands/__constants';
import { CommandDoc } from 'src/types';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

export async function appendCommand(name: string, newInstruction: string): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    Feedback.title(`Append Instruction to Command: ${name}`);

    commandDoc.instructions.push(newInstruction);

    await db.put(commandDoc);

    Feedback.success(`Instruction "${newInstruction}" added to "${name}".`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${name}" not found`);
    } else {
      Feedback.error(`Failed to add instruction to "${name}": ${error.message}`);
    }
  }
}
