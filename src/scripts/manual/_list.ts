import db from '../../db';

export async function listCommand() {
  try {
    const result = await db.allDocs({ include_docs: false });

    const manuals = result.rows
      .filter((row) => row.id.startsWith('manual:'))
      .map((row) => row.id.replace('manual:', ''));

    if (manuals.length === 0) {
      console.log('📭 No manuals found.');
      return;
    }

    console.log('📚 Available Manuals:\n');
    manuals.forEach((name) => console.log(`- ${name}`));
  } catch (error: any) {
    console.error(`❌ Failed to list manuals: ${error.message}`);
  }
}
