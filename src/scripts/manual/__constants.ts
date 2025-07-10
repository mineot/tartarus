export const MANUAL_PREFIX = 'manual:';
export const MANUAL_EDITOR_CONFIG_ID = 'config:editor';

export function getManualTempFile(name: string): string {
  return `tartarus_manual_${name}.txt`;
}
