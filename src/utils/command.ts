import { error } from './error';
import kleur from 'kleur';

const DONE_COLOR = kleur.green;
const DONE_ICON = '✅';

export interface Command {
  validation: (args: any) => Promise<void>;
  execution: (args: any) => Promise<string>;
}

export async function command(args: any, cmd: Command): Promise<void> {
  const { validation, execution } = cmd;

  try {
    await validation(args);
    const result = await execution(args);

    if (result) {
      console.log(DONE_COLOR(`${DONE_ICON} ${result}`));
    }
  } catch (err: any) {
    error(err);
  }
}
