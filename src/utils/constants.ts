export const COMMAND_PREFIX = 'cmd:';
export const MANUAL_EDITOR_CONFIG_ID = 'config:editor';
export const MANUAL_PREFIX = 'manual:';

export function getManualTempFile(name: string): string {
  return `tartarus_manual_${name}.txt`;
}
