import { Command } from 'commander';

import { registerCMD } from 'src/scripts/commands';
import { registerDatabase } from 'src/scripts/database';
import { registerRun } from 'src/scripts/run';
import { registerManual } from 'src/scripts/manual';

const program = new Command();

program
  .name('tartarus')
  .version('1.1.1')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCMD(program);
registerDatabase(program);
registerRun(program);
registerManual(program);

program.parse();
