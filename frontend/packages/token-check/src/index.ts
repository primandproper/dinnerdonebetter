import { readdirSync, readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

// The apps' own styles read @primandproper/platform-ui's --pp-* tokens, and the library's theme is the only place
// they are defined: a token the library renames or drops would leave those styles unset without failing the build.
// Both apps run this from their own tests, so the check lives here once instead of drifting in two copies.

export interface TokenRead {
  token: string;
  path: string;
}

// undefinedTokenReads scans every .svelte and .css file under srcDir for var(--pp-*) reads and returns the ones the
// platform-ui theme does not define. The library is resolved from srcDir, so each app is checked against the
// version it installed. It throws if the scan finds no reads at all, since that means srcDir is wrong and an
// empty result would pass vacuously.
export function undefinedTokenReads(srcDir: URL): TokenRead[] {
  const dir = fileURLToPath(srcDir);
  const defined = definedTokens(dir);
  const reads = tokenReads(dir);
  if (reads.size === 0) {
    throw new Error(`no var(--pp-*) reads found under ${dir}`);
  }
  return [...reads].filter(([token]) => !defined.has(token)).map(([token, path]) => ({ token, path }));
}

function definedTokens(fromDir: string): Set<string> {
  const require = createRequire(join(fromDir, 'noop.js'));
  const libraryDir = dirname(require.resolve('@primandproper/platform-ui/package.json'));
  const theme = readFileSync(join(libraryDir, 'dist', 'theme.css'), 'utf8');
  return new Set([...theme.matchAll(/(--pp-[a-z0-9-]+)\s*:/g)].map((m) => m[1]));
}

function tokenReads(dir: string): Map<string, string> {
  const reads = new Map<string, string>();
  for (const entry of readdirSync(dir, { withFileTypes: true, recursive: true })) {
    if (!entry.isFile() || !/\.(svelte|css)$/.test(entry.name)) continue;
    const path = join(entry.parentPath, entry.name);
    for (const m of readFileSync(path, 'utf8').matchAll(/var\(\s*(--pp-[a-z0-9-]+)/g)) {
      reads.set(m[1], path);
    }
  }
  return reads;
}
