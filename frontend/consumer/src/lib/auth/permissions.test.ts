import { randomUUID } from 'node:crypto';
import { describe, it, expect } from 'vitest';
import { holds, Permission } from './permissions';

const all = Object.values(Permission);

function pick<T>(values: readonly T[]): T {
  return values[Math.floor(Math.random() * values.length)];
}

describe('holds', () => {
  it('is true for a permission the server listed', () => {
    const permission = pick(all);

    expect(holds({ permissions: { permissions: [`other.${randomUUID()}`, permission] } }, permission)).toBe(true);
  });

  it('is false for one it did not list, whatever else was', () => {
    const permission = pick(all);
    const others = all.filter((p) => p !== permission);

    expect(holds({ permissions: { permissions: others } }, permission)).toBe(false);
  });

  it('is false for anything when the server stated an empty list', () => {
    expect(holds({ permissions: { permissions: [] } }, pick(all))).toBe(false);
  });

  it('leaves the call to the server when it stated no permissions at all', () => {
    expect(holds({ permissions: undefined }, pick(all))).toBe(true);
  });
});
