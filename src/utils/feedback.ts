import kleur from 'kleur';

export class Feedback {
  static success(message: string) {
    console.log(kleur.green(`✅ ${message}`));
  }

  static error(message: string) {
    console.error(`❌ ${kleur.bgRed().white(message)}`);
  }

  static warn(message: string) {
    console.warn(`⚠️ ${kleur.bgYellow().white(message)}`);
  }

  static info(message: string) {
    console.log(kleur.cyan(`ℹ️ ${message}`));
  }

  static title(message: string) {
    console.log(kleur.bold().white(`\n${message}\n`));
  }

  static text(message: string) {
    console.log(message);
  }

  static enphased(message: string) {
    console.log(kleur.italic().gray(message));
  }

  static notAllowed(message: string) {
    console.log(kleur.magenta(`⛔ ${message}`));
  }

  static exists(message: string) {
    console.log(`📬 ${kleur.magenta(message)}`);
  }

  static notFound(message: string) {
    console.log(`📭 ${kleur.bgMagenta().white(message)}`);
  }

  static item(message: string) {
    console.log(`➡️ ${message}`);
  }

  static divider() {
    console.log(kleur.white('─'.repeat(40)));
  }

  static break() {
    console.log('');
  }
}
