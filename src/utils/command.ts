import { Command } from 'commander';
import kleur from 'kleur';

export type OperatorReturn = Promise<string | null>;
export type Args = string[];

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
  operation: (args: Args) => OperatorReturn;
}

export async function command(args: Args, setting: Setting): Promise<void> {
  const { referenceName, operation } = setting;
  const firstArg = args[0];
  const breakLine = '\n';
  const lineRepetition = '─'.repeat(50);

  try {
    console.log(kleur.bold().white(`Execute the '${referenceName}' command:`));
    console.log(kleur.white(lineRepetition.concat(breakLine)));

    if (args.length === 0) {
      throw { message: 'No arguments found' };
    }

    const result: string | null = await operation(args);

    if (result) {
      console.log(result);
      console.log(kleur.white(breakLine.concat(lineRepetition)));
    }

    console.log(kleur.green(`Command '${referenceName}' completed successfully`));
  } catch (error: any) {
    if (error.status === 404) {
      console.log(
        kleur
          .bgMagenta()
          .white(
            `An error occurred while executing the '${referenceName}' command: '${firstArg}' was not found.`
          )
      );
    } else {
      console.error(
        kleur.bgRed().white(`Error during execution the command ${referenceName}: ${error.message}`)
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
