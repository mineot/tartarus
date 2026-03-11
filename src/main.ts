import { databaseMigrator } from '@database';
import { runCommand } from '@commands';

async function bootstrap(): Promise<void> {
  await databaseMigrator;
  await runCommand();
}

bootstrap().catch((error: unknown) => {
  console.error('Failed to bootstrap application:', error);
  process.exit(1);
});
