import { undefinedTokenReads } from '@dinnerdonebetter/token-check';
import { describe, expect, it } from 'vitest';

describe('platform-ui tokens', () => {
  it('that this app reads are all ones the library defines', () => {
    expect(undefinedTokenReads(new URL('.', import.meta.url))).toEqual([]);
  });
});
