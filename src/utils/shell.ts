import { execSync, type ExecSyncOptions } from 'node:child_process';
import fs from 'node:fs';

function exists(p?: string) {
  return !!p && fs.existsSync(p);
}

async function resolveShell(): Promise<string> {
  const override = process.env.TARTARUS_SHELL;
  if (override && exists(override)) return override;

  if (process.platform === 'win32') {
    const candidates = [
      process.env.COMSPEC,
      'C:\\Program Files\\Git\\bin\\bash.exe',
      'C:\\Program Files (x86)\\Git\\bin\\bash.exe',
      'C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe',
      'cmd.exe',
    ];

    for (const c of candidates) {
      if (exists(c)) {
        return c!;
      }
    }

    return 'cmd.exe';
  }

  return exists('/bin/bash') ? '/bin/bash' : '/bin/sh';
}

export async function execInstruction(
  instruction: string,
  opts: ExecSyncOptions = {}
): Promise<void> {
  execSync(instruction, { stdio: 'inherit', shell: await resolveShell(), ...opts });
}
