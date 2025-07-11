import db from '../../db';
import fs from 'fs';
import path from 'path';
import { Feedback } from '../../utils/feedback';

export async function importCommand(filePath: string) {
  try {
    const resolvedPath = path.resolve(filePath);

    if (!fs.existsSync(resolvedPath)) {
      Feedback.notFound(`File not found: ${resolvedPath}`);
      return;
    }

    const fileContent = fs.readFileSync(resolvedPath, 'utf-8');
    const docs = JSON.parse(fileContent);

    if (!Array.isArray(docs)) {
      Feedback.error(`Invalid backup format. Expected an array of documents.`);
      return;
    }

    const cleanDocs = docs.map((doc) => {
      const { _rev, ...rest } = doc;
      return rest;
    });

    await db.bulkDocs(cleanDocs);
    Feedback.success(`Backup imported from: ${resolvedPath}`);
  } catch (error: any) {
    Feedback.error(`Import failed: ${error.message}`);
  }
}
