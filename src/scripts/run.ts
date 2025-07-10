import { Command } from 'commander';
import { CommandDoc } from '../types';
import { execSync } from 'child_process';
import { Feedback } from '../utils/feedback';
import db from '../db';

/**
 * Executes all instructions registered for the given command.
 * @param {string} name The name of the command to execute.
 * @returns {Promise<void>} A promise that resolves when the execution is completed.
 */
async function runCommand(name: string): Promise<void> {
  try {
    // Retrieve the command document by its name.
    const commandDoc = (await db.get(name)) as CommandDoc;

    // Iterate over each instruction in the command document.
    for (const [index, instruction] of commandDoc.instructions.entries()) {
      // Print the instruction with its index.
      Feedback.message(`\n  [${index + 1}] ${instruction}`);

      // Execute the instruction using the `execSync` function.
      // The `{ stdio: 'inherit', shell: '/bin/bash' }` options allow the command
      // to inherit the parent process's standard input/output/error streams and
      // use the Bash shell to execute the instruction.
      execSync(instruction, { stdio: 'inherit', shell: '/bin/bash' });
    }
  } catch (error: any) {
    // Handle the case where the command is not found.
    if (error.status === 404) {
      Feedback.error(`Command "${name}" not found`);
      return;
    }

    // Handle other errors.
    Feedback.error(`Error during execution: ${error.message}`);
  }
}

/**
 * Registers the `exec` command group.
 * @param {Command} program The Commander program to register the command group with.
 */
export function registerRunCommand(program: Command): void {
  /**
   * Executes all instructions registered for the given command.
   * @param {string} commandName The name of the command to execute.
   * @returns {Promise<void>} A promise that resolves when the execution is completed.
   */
  program
    .command('run <commandName>')
    .description('Execute all instructions registered for the given command')
    .action(runCommand);
}
