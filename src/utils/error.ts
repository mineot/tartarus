import kleur from 'kleur';

const ERROR_COLOR = kleur.red;
const ERROR_ICON = '🛑';
const WARNING_COLOR = kleur.magenta;
const WARNING_ICON = '⚠️';

export type Error = {
  message: string;
  status?: 404 | 500;
  referenceName?: string;
  commandName?: string;
};

export const FailThrow = (message: string) => {
  throw { status: 500, message };
};

export async function error(error: Error): Promise<void> {
  if (!error) {
    return;
  }

  const { status } = error;
  let { message, referenceName, commandName } = error;

  message = message ?? 'No Message';
  referenceName = referenceName ?? 'No Reference Name';
  commandName = commandName ?? 'No Command Name';

  if (status === 404) {
    message = `A problem occurred while running the command "${referenceName}": "${commandName}" was not found.`;
  }

  if (status) {
    console.log(WARNING_COLOR(`\n${WARNING_ICON} ${message}`));
  } else {
    console.log(ERROR_COLOR(`\n${ERROR_ICON} ${message}`));
  }
}
