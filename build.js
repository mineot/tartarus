// build.js
const esbuild = require('esbuild');

esbuild
  .build({
    entryPoints: ['src/cli.ts'],
    bundle: true,
    minify: true,
    platform: 'node',
    target: 'node18',
    outfile: 'dist/tartarus.js',
    external: ['pouchdb'],
    banner: {
      js: '#!/usr/bin/env node',
    },
  })
  .then(() => {
    console.log('✅ Build complete: dist/tartarus.js');
  })
  .catch(() => process.exit(1));
