import { Command } from 'commander';
import { meta } from 'src/meta';
// import { command, register } from './utils/command';

import { registerCMD } from 'src/scripts/commands';
import { registerDatabase } from 'src/scripts/database';
import { registerRun } from 'src/scripts/run';
import { registerManual } from 'src/scripts/manual';

const program = new Command();

// register({
//   program,
//   commandName: 'test',
//   commandDescription: `Command Description`,
//   commandHelp: [
//     { name: 'name', description: 'Description Name' },
//     { name: 'other', description: 'Description Other' },
//   ],
//   commandInstance: (args) =>
//     command(args, {
//       referenceName: 'test',
//       operation: (args) => Promise.resolve(console.log(args)),
//     }),
// });

program.name(meta.name).version(meta.version).description(meta.description);
registerDatabase(program);
registerManual(program);
registerCMD(program);
registerRun(program);
program.parse();
