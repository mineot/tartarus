import { Command } from 'commander';

export type Argument = {
  name: string;
  description: string;
  required?: boolean;
  variadic?: boolean;
};

export interface Register {
  program: Command;
  commandName: string;
  commandArguments: Argument[];
  commandDescription: string;
  commandInstance: (args: any) => Promise<void>;
}

export function register(param: Register): void {
  const { commandName, commandDescription, commandInstance, commandArguments } = param;

  const program = param.program
    .allowUnknownOption(false)
    .command(commandName)
    .description(commandDescription)
    .action((...received: any[]) => {
      const list = received.slice(0, -2);
      const processed = param.commandArguments
        .map((argument, index) => ({
          [argument.name]: list[index],
        }))
        .reduce((acc, item) => ({ ...acc, ...item }), {});
      return commandInstance(processed);
    });

  commandArguments.forEach((argument) => {
    let token = argument.name;
    token = argument.variadic ? `${token}...` : token;
    token = argument.required ? `<${token}>` : `[${token}]`;
    program.argument(token, argument.description);
  });
}
