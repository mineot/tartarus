import { Command } from 'commander';
import kleur from 'kleur';

export interface Register {
  program: Command;
  commandName: string;
  commandDescription: string;
  commandInstance: (args: string[]) => Promise<void>;
}

export interface Setting {
  referenceName: string;
  operation: (args: string[]) => Promise<void>;
}

export async function command(args: string[], setting: Setting): Promise<void> {
  const { referenceName, operation } = setting;
  const firstArg = args[0];
  const breakLine = '\n\n';
  const lineRepetition = '─'.repeat(50);

  try {
    console.log(kleur.bold().white(`Execute the '${referenceName}' command:`));
    console.log(kleur.white(lineRepetition.concat(breakLine)));

    if (args.length === 0) {
      throw { message: 'No arguments found' };
    }

    await operation(args);

    console.log(kleur.white(breakLine.concat(lineRepetition)));
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
    .command(`${param.commandName} [args...]`)
    .description(param.commandDescription)
    .action((args) => param.commandInstance(args));
}
