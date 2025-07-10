import { ConfigDoc } from '../../types';
import { Feedback } from '../../utils/feedback';
import { MANUAL_EDITOR_CONFIG_ID } from './__constants';
import db from '../../db';

/**
 * Sets the default editor application.
 * @param {string} editor - The name of the editor (e.g. nano, vim, notepad).
 * @returns {Promise<void>} - A promise that resolves when the editor is set.
 */
export async function setEditorApp(editor: string): Promise<void> {
  /**
   * Validate the editor name input
   */
  if (!editor || !editor.trim()) {
    Feedback.warn(
      'You must provide a valid editor name (e.g. nano, vim, notepad).\nSee the help for more information.'
    );
    return;
  }

  /**
   * Construct the configuration document
   */
  const config: ConfigDoc = {
    _id: MANUAL_EDITOR_CONFIG_ID,
    data: editor.trim(),
    updatedAt: new Date().toISOString(),
  };

  try {
    /**
     * Update the existing configuration
     */
    const existing = await db.get(MANUAL_EDITOR_CONFIG_ID);
    await db.put({ ...existing, ...config });
    Feedback.success(`Editor updated to "${editor}".`);
  } catch {
    /**
     * Create a new configuration if it does not exist
     */
    await db.put(config);
    Feedback.success(`Editor set to "${editor}".`);
  }
}

/**
 * Shows the default editor application.
 * @returns {Promise<void>} - A promise that resolves when the current editor is displayed.
 */
export async function showEditorApp(): Promise<void> {
  /**
   * Retrieve the current editor configuration
   */
  const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;

  if (config.data && typeof config.data === 'string') {
    /**
     * Display the current editor
     */
    Feedback.message(`Current editor: ${config.data}`);
  } else {
    /**
     * Display a message if no editor is configured
     */
    Feedback.info('No editor configured. Use: tartarus man app <editor>');
  }
}

/**
 * Retrieves the default editor application from the database.
 * If the editor is not configured, the function throws an error and exits the process.
 * @returns {Promise<string>} - A promise that resolves with the default editor application.
 */
export async function getEditorAppOrFail(): Promise<string> {
  try {
    /**
     * Retrieve the current editor configuration
     */
    const config = (await db.get(MANUAL_EDITOR_CONFIG_ID)) as ConfigDoc;

    if (!config.data || typeof config.data !== 'string') {
      /**
       * Throw an error if the editor is not configured
       */
      throw new Error();
    }

    /**
     * Return the current editor
     */
    return config.data;
  } catch {
    /**
     * Display an error message if the editor is not configured
     */
    Feedback.error('No editor configured. Use: tartarus man set_editor <editor>');
    /**
     * Exit the process with a non-zero code to indicate failure
     */
    process.exit(1);
  }
}
