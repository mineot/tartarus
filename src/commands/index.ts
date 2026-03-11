import { program } from 'commander';

export async function runCommand() {
  program.helpOption(false);

  program.name('tartarus').description('Tartarus CLI');

  program
    .command('cmd')
    .description('Menager Commands')
    .action(async () => {
      console.log('Menager all commands...');
    });

  program.parse();
}
