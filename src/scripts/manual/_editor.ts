import { ConfigDoc } from 'src/types';
import { Feedback } from 'src/utils/feedback';
import { MANUAL_EDITOR_CONFIG_ID } from 'src/scripts/manual/__constants';
import db from 'src/db';

export async function setEditorApp(editor: string): Promise<void> {
  if (!editor || !editor.trim()) {
    Feedback.warn(
      'You must provide a valid editor name (e.g. nano, vim, notepad).\nSee the help for more information.'
    );
    return;
  }

  const config: ConfigDoc = {
    _id: MANUAL_EDITOR_CONFIG_ID,
    data: editor.trim(),
    updatedAt: new Date().toISOString(),
  };

  try {
    const existing = await db.get(MANUAL_EDITOR_CONFIG_ID);
    await db.put({ ...existing, ...config });
    Feedback.success(`Editor updated to "${editor}".`);
  } catch {
    await db.put(config);
    Feedback.success(`Editor set to "${editor}".`);
  }
}

export async function showEditorApp(): Promise<void> {
  const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;

  if (config.data && typeof config.data === 'string') {
    Feedback.text(`Current editor: ${config.data}`);
  } else {
    Feedback.notFound('No editor configured. Use: tartarus man app <editor>');
  }
}
export async function getEditorAppOrFail(): Promise<string> {
  try {
    const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;

    if (!config.data || typeof config.data !== 'string') {
      throw new Error();
    }

    return config.data;
  } catch {
    Feedback.error('No editor configured. Use: tartarus man set_editor <editor>');
    process.exit(1);
  }
}
