import { Command } from 'commander';
import { meta } from 'src/meta';

import { registerCMD } from 'src/scripts/commands';
import { registerDatabase } from 'src/scripts/database';
import { registerRun } from 'src/scripts/run';
import { registerManual } from 'src/scripts/manual';

const program = new Command();

program.name(meta.name).version(meta.version).description(meta.description);
registerDatabase(program);
registerManual(program);
registerCMD(program);
registerRun(program);
program.parse();
