import { Command } from 'commander';

import { registerCmdGroup } from './scripts/cmd';
import { registerExecCommand } from './scripts/exec';
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
registerExecCommand(program);
registerDatabaseGroup(program);
registerDocGroup(program);

program.parse();
