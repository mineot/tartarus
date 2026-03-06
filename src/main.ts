import { database, databaseMigrator } from '@database';

async function bootstrap(): Promise<void> {
  await databaseMigrator;
  console.log('Tartarus');
  await database.selectFrom('migrations').select('name').execute();
}

bootstrap().catch((error: unknown) => {
  console.error('Failed to bootstrap application:', error);
  process.exit(1);
});
