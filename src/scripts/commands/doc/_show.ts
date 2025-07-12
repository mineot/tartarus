import { COMMAND_PREFIX } from 'src/scripts/commands/__constants';
import { CommandDoc } from 'src/types';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

export async function showDocCommand(name?: string) {
  try {
    if (name) {
      const prefixName = `${COMMAND_PREFIX}${name}`;
      const doc = (await db.get(prefixName)) as CommandDoc;

      if (!doc.description) {
        Feedback.warn(`Command "${name}" has no description.`);
        return;
      }

      Feedback.title(`"${name}" command documentation:`);
      Feedback.text(doc.description);
    }

    const allDocs = await db.allDocs({ include_docs: true });

    const described = allDocs.rows
      .filter((row) => row.id.startsWith(COMMAND_PREFIX))
      .map((row) => row.doc)
      .filter((doc) => (doc as CommandDoc)?.description);

    if (described.length === 0) {
      Feedback.notFound('No documented commands found.');
      return;
    }

    for (const doc of described) {
      Feedback.title(`"${doc?._id.replace(COMMAND_PREFIX, '')}" command documentation:`);
      Feedback.text(`${(doc as CommandDoc)?.description}\n`);
    }
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${name}" not found`);
    } else {
      Feedback.error(`Failed to show documentation for "${name}": ${error.message}`);
    }
  }
}
