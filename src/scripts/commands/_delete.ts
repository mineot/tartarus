import { COMMAND_PREFIX } from 'src/utils/constants';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

export async function deleteCommand(name: string) {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const command = await db.get(prefixName);
    await db.remove(command);
    Feedback.success(`Command "${name}" removed successfully`);
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${name}" not found`);
    } else {
      Feedback.error(`Failed to remove command "${name}": ${error.message}`);
    }
  }
}
