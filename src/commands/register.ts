import db from '@/db';

export default async function register(name: string, command: string) {
  try {
    await db.put({ _id: name, command });
    console.log(`✅ Comando "${name}" registrado.`);
  } catch (err: any) {
    if (err.status === 409) {
      console.error('❌ Já existe um comando com esse nome.');
    } else {
      console.error('❌ Erro ao registrar:', err.message);
    }
  }
}
