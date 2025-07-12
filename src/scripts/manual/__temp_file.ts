import { Feedback } from 'src/utils/feedback';
import { getEditorAppOrFail } from 'src/scripts/manual/_editor';
import { getManualTempFile } from 'src/scripts/manual/__constants';
import { spawnSync } from 'child_process';
import fs from 'fs';
import os from 'os';
import path from 'path';

export async function tempFile(name: string, initContent: string): Promise<string | null> {
  const editor = await getEditorAppOrFail();

  const tempPath = path.join(os.tmpdir(), getManualTempFile(name));
  fs.writeFileSync(tempPath, initContent);

  spawnSync(editor, [tempPath], { stdio: 'inherit' });

  const content = fs.readFileSync(tempPath, 'utf-8').trim();
  fs.unlinkSync(tempPath);

  if (!content) {
    Feedback.warn('No content written. Manual not saved.');
    return null;
  }

  return content;
}
