import { Command } from 'commander';
import { createCommand } from './_create';
import { editCommand } from './_edit';
import { setEditorApp, showEditorApp } from './_editor';
import { showCommand } from './_show';
import { deleteCommand } from './_delete';
import { listCommand } from './_list';

export function registerManual(program: Command): void {
  const man = program.command('man').description('Manual management commands');

  man
    .command('set_editor')
    .argument('<editorName>', 'Editor name (e.g. nano, vim, notepad)')
    .description('Set the default editor')
    .action(setEditorApp);

  man.command('show_editor').description('Show the default editor').action(showEditorApp);

  man
    .command('create')
    .argument('<manualName>', 'Name of the new manual')
    .description('Create a new manual')
    .action(createCommand);

  man
    .command('edit')
    .argument('<manualName>', 'Name of the manual to edit')
    .description('Edit an existing manual')
    .action(editCommand);

  man
    .command('delete')
    .argument('<manualName>', 'Name of the manual to delete')
    .description('Delete a manual')
    .action(deleteCommand);

  man
    .command('show')
    .argument('<manualName>', 'Name of the manual to show')
    .description('Show the content of a manual')
    .action(showCommand);

  man.command('list').description('List available manuals').action(listCommand);
}
