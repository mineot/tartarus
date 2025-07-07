#!/usr/bin/env node
import { Command } from 'commander';
import register from '@/commands/register';
import execCommand from '@/commands/execute';
import listCommands from '@/commands/list';
import removeCommand from '@/commands/remove';
import exportCommands from '@/commands/export';
import importCommands from '@/commands/import';

const program = new Command();

program.name('tartarus').description('CLI para comandos nomeados').version('1.0.0');

program.command('register').argument('<name>').argument('<cmd>').action(register);
program.command('exec').argument('<name>').action(execCommand);
program.command('list').action(listCommands);
program.command('remove').argument('<name>').action(removeCommand);
program.command('export').action(exportCommands);
program.command('import').action(importCommands);

program.parse();
