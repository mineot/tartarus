import { Command } from 'commander';
import { registerManCreate } from './_create';
import { registerManDelete } from './_delete';
import { registerManEdit } from './_edit';
import { registerManList } from './_list';
import { registerManSetEditor } from './_set_editor';
import { registerManShow } from './_show';
import { registerManShowEditor } from './_show_editor';

export function registerManual(program: Command): void {
  const man = program.command('man').description('Manual management.');
  registerManSetEditor(man);
  registerManShowEditor(man);
  registerManCreate(man);
  registerManEdit(man);
  registerManDelete(man);
  registerManShow(man);
  registerManList(man);
}
