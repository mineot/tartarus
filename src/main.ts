import { Command } from 'commander';
import { meta } from './core/meta';

import { registerCMD } from 'src/scripts/commands';
import { registerDatabase } from 'src/scripts/database';
import { registerRun } from 'src/scripts/run';
import { registerManual } from 'src/scripts/manual';

const program = new Command();

program.name(meta.name).version(meta.version).description(meta.description);
registerCMD(program);
registerRun(program);
registerManual(program);
registerDatabase(program);
program.parse();
