import { Feedback } from 'src/utils/feedback';
import db from 'src/db';
import fs from 'fs';
import path from 'path';

export async function exportCommand(filePath: string) {
  try {
    const resolvedPath = path.resolve(filePath);
    const result = await db.allDocs({ include_docs: true });

    const allDocs = result.rows.map((row) => row.doc).filter(Boolean);
    fs.writeFileSync(resolvedPath, JSON.stringify(allDocs, null, 2));

    Feedback.success(`Backup exported to ${resolvedPath}`);
  } catch (error: any) {
    Feedback.error(`Export failed: ${error.message}`);
  }
}
