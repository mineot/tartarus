import kleur from 'kleur';

const DONE_COLOR = kleur.green;
const DONE_ICON = '✅';
const TITLE_COLOR = kleur.bold().gray;
const TITLE_COLOR_DESCRIPTION = kleur.reset().italic().grey;
const TITLE_COLOR_ITEM = kleur.white;
const TITLE_ICON = '📃';
const TITLE_ICON_ITEM = '➡️';

export type OutputParam = {
  title: string;
  description?: string;
  list: string[];
  underlineTitle?: boolean;
};

export async function done(message: string) {
  console.log(DONE_COLOR(`\n${DONE_ICON} ${message}`));
}

export async function output(param: OutputParam) {
  const { title, description, list, underlineTitle } = param;
  let text = '';

  if (description) {
    text = TITLE_COLOR(`\n${TITLE_ICON} ${title}: ${TITLE_COLOR_DESCRIPTION(description)}`);
  } else {
    text = TITLE_COLOR(`\n${TITLE_ICON} ${title}`);
  }

  console.log(text);

  if (underlineTitle) {
    console.log(TITLE_COLOR('-'.repeat(text.length)));
  }

  list.forEach((item) => {
    console.log(TITLE_COLOR_ITEM(`${TITLE_ICON_ITEM} ${item}`));
  });
}
