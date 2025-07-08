import db from '../db';

export default async function register(name: string, command: string) {
  try {
    await db.put({ _id: name, instructions: [command] });
    console.log(`✅ Command "${name}" registered with 1 instruction.`);
  } catch (err: any) {
    if (err.status === 409) {
      console.error('❌ A command with this name already exists.');
    } else {
      console.error('❌ Registration error:', err.message);
    }
  }
}
