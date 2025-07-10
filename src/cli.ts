import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';
import { registerDatabaseGroup } from './scripts/database';
import { registerRunCommand } from './scripts/run';

const program = new Command();

program
  .name('tartarus')
  .version('1.0.0')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCmdGroup(program);
registerDatabaseGroup(program);
registerRunCommand(program);

program.parse();
