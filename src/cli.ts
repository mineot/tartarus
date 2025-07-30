import { Command } from 'commander';
import { meta } from 'src/meta';
import { command, register } from './utils/command';

import { registerCMD } from 'src/scripts/commands';
import { registerDatabase } from 'src/scripts/database';
import { registerRun } from 'src/scripts/run';
import { registerManual } from 'src/scripts/manual';

const program = new Command();

register({
  commandName: 'test',
  commandDescription: 'Test Description.',
  program,
  commandInstance: (args) =>
    command(args, {
      referenceName: 'test',
      operation: (args) => Promise.resolve(console.log(args)),
    }),
});

program.name(meta.name).version(meta.version).description(meta.description);
registerDatabase(program);
registerManual(program);
registerCMD(program);
registerRun(program);
program.parse();
