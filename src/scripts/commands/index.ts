import { Command } from 'commander';
import { registerCmdAppend } from './_append';
import { registerCmdCreate } from './_create';
import { registerCmdDelete } from './_delete';
import { registerCmdList } from './_list';
import { registerCmdShow } from './_show';
import { registerCmdSubtract } from './_subtract';
import { registerDocumentation } from './doc';

export function registerCMD(program: Command): void {
  const cmd = program.command('cmd').description('Command management');
  registerCmdCreate(cmd);
  registerCmdAppend(cmd);
  registerCmdSubtract(cmd);
  registerCmdDelete(cmd);
  registerCmdShow(cmd);
  registerCmdList(cmd);
  registerDocumentation(cmd);
}
