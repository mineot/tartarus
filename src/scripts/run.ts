import { Command } from 'commander';
import { COMMAND_PREFIX } from 'src/scripts/commands/__constants';
import { CommandDoc } from 'src/types';
import { execSync } from 'child_process';
import { Feedback } from 'src/utils/feedback';
import db from 'src/db';

async function runCommand(name: string): Promise<void> {
  try {
    const prefixName = `${COMMAND_PREFIX}${name}`;
    const commandDoc = (await db.get(prefixName)) as CommandDoc;

    for (const [index, instruction] of commandDoc.instructions.entries()) {
      Feedback.item(`\n  [${index + 1}] - ${instruction}`);
      execSync(instruction, { stdio: 'inherit', shell: '/bin/bash' });
    }
  } catch (error: any) {
    if (error.status === 404) {
      Feedback.notFound(`Command "${name}" not found`);
      return;
    }

    Feedback.error(`Error during execution: ${error.message}`);
  }
}

export function registerRun(program: Command): void {
  program
    .command('run [commandName]')
    .description('Execute all instructions registered for the given command')
    .action(runCommand);
}
