import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';
import { registerRunCommand } from './scripts/run';
import { registerDatabaseGroup } from './scripts/database';
import { registerDocGroup } from './scripts/doc';

const program = new Command();

program
  .name('tartarus')
  .version('1.0.0')
  .description(
    'Register, organize, and execute named shell command groups with descriptions and backups.'
  );

registerCmdGroup(program);
registerDocGroup(program);
registerDatabaseGroup(program);
registerRunCommand(program);

program.parse();
