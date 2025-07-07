import db from '../db';

export default async function registerCommand(name: string, command: string) {
  try {
    await db.put({ _id: name, command });
    console.log(`✅ Command "${name}" registered.`);
  } catch (err: any) {
    if (err.status === 409) {
      console.error('❌ Command with this name already exists.');
    } else {
      console.error('❌ Error registering:', err.message);
    }
  }
}
