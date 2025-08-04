import { Command } from 'commander';
import { registerAddDocumentation } from './_add';
import { registerListDocumentation } from './_list';
import { registerRemoveDocumentation } from './_remove';
import { registerShowDocumentation } from './_show';

export function registerDocumentation(program: Command): void {
  const doc = program.command('doc').description('Documentation management');
  registerAddDocumentation(doc);
  registerRemoveDocumentation(doc);
  registerListDocumentation(doc);
  registerShowDocumentation(doc);
}
