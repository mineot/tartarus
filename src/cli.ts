import { Command } from 'commander';

import { registerCMD } from './scripts/commands';
import { registerDatabase } from './scripts/database';
import { registerRun } from './scripts/run';

const program = new Command();

program
  .name('tartarus')
  .version('1.0.0')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCMD(program);
registerDatabase(program);
registerRun(program);

program.parse();
