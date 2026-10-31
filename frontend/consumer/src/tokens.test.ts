import { readdirSync, readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// This app's own styles read @primandproper/platform-ui's --pp-* tokens, and the library's theme is the only place
// they are defined: a token the library renames or drops would leave those styles unset without failing the build.

const srcDir = dirname(fileURLToPath(import.meta.url));
const libraryDir = dirname(createRequire(import.meta.url).resolve('@primandproper/platform-ui/package.json'));

function definedTokens(): Set<string> {
  const theme = readFileSync(join(libraryDir, 'dist', 'theme.css'), 'utf8');
  return new Set([...theme.matchAll(/(--pp-[a-z0-9-]+)\s*:/g)].map((m) => m[1]));
}

function readTokens(dir: string): Map<string, string> {
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

describe('platform-ui tokens', () => {
  it('that this app reads are all ones the library defines', () => {
    const defined = definedTokens();
    const reads = readTokens(srcDir);

    expect(reads.size).toBeGreaterThan(0);
    const undefinedReads = [...reads].filter(([token]) => !defined.has(token));
    expect(undefinedReads).toEqual([]);
  });
});
