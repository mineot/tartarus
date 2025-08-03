import { ConfigDoc } from 'src/core/types';
import { getManualTempFile, MANUAL_EDITOR_CONFIG_ID } from '../../utils/constants';
import { spawnSync } from 'child_process';
import db from 'src/core/db';
import fs from 'fs';
import os from 'os';
import path from 'path';

export async function getEditorAppOrFail(): Promise<string> {
  const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;
  if (!config.data || typeof config.data !== 'string') {
    throw 'No editor configured. Use: tartarus man set_editor <editor>';
  }
  return config.data;
}

export async function tempFile(name: string, initContent: string): Promise<string | null> {
  try {
    const editor = await getEditorAppOrFail();

    const tempPath = path.join(os.tmpdir(), getManualTempFile(name));
    fs.writeFileSync(tempPath, initContent);

    spawnSync(editor, [tempPath], { stdio: 'inherit' });

    const content = fs.readFileSync(tempPath, 'utf-8').trim();
    fs.unlinkSync(tempPath);

    if (!content) {
      throw 'No content written. Manual not saved.';
    }

    return content;
  } catch (error: any) {
    throw error;
  }
}
