import { Command } from 'commander';
import kleur from 'kleur';

export type Args = string[];
export type OperationReturn = Promise<string | null>;
export type Result = string | null;
export type ValidationReturn = Promise<void>;

export interface Help {
  structure: { name: string; description: string }[];
  example: string;
}

export interface Register {
  program: Command;
  commandName: string;
  commandDescription: string;
  commandHelp: Help;
  commandInstance: (args: Args) => Promise<void>;
}

export interface Setting {
  referenceName: string;
  noArguments?: boolean;
  validation: (args: Args) => ValidationReturn;
  operation: (args: Args) => OperationReturn;
}

export const FailThrow = (message: string) => {
  throw { status: 500, message };
};

export const TitledText = (title: string, text: string) => {
  console.log(`${kleur.bold(kleur.white(title))}: ${kleur.white(text)}\n`);
};

export async function command(args: Args, setting: Setting): Promise<void> {
  const { referenceName, operation, validation, noArguments } = setting;

  try {
    if (!noArguments && args.length === 0) {
      FailThrow('Arguments not found');
    }

    await validation(args);

    const result: Result = await operation(args);

    if (result) {
      console.log(kleur.green(`✅ ${result}`));
    }
  } catch (error: any) {
    if (error.status === 404) {
      const [firstArg] = args;

      console.log(
        kleur.magenta(
          `⚠️ A problem occurred while running the command "${referenceName}": ❓ '${firstArg}' was not found.`
        )
      );
    } else if (error.status === 500) {
      console.log(kleur.magenta(`⚠️ ${error.message}`));
    } else {
      console.error(
        kleur.red(`🛑 Error during execution the command ${referenceName}: ${error.message}`)
      );
    }
  }
}

export function register(param: Register): void {
  param.program
    .command(`${param.commandName}`)
    .argument('[args...]', 'Arguments')
    .description(param.commandDescription)
    .allowUnknownOption(false)
    .addHelpText(
      'after',
      `
The 'args' arguments must follow this structure:

${param.commandHelp.structure.map((help) => `- <${help.name}>: ${help.description}`).join('\n')}

Example: ${param.commandHelp.example}
`
    )
    .action((args) => param.commandInstance(args));
}
