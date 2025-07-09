import kleur from 'kleur';

export const Feedback = {
  success: (message: string) => {
    console.log(kleur.green(`✅ ${message}`));
  },
  error: (message: string) => {
    console.error(kleur.red(`❌ ${message}'`));
  },
  notAllowed: (message: string) => {
    console.log(kleur.yellow(`⛔ ${message}`));
  },
  warn: (message: string) => {
    console.warn(kleur.yellow(`⚠️ ${message}`));
  },
  info: (message: string) => {
    console.log(kleur.cyan(`ℹ️ ${message}`));
  },
  message: (message: string) => {
    console.log(message);
  },
};
