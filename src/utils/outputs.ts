import kleur from 'kleur';

const DONE_ICON = '✅';
const DONE_COLOR = kleur.green;

export async function done(message: string) {
  console.log(DONE_COLOR(`\n${DONE_ICON} ${message}`));
}
