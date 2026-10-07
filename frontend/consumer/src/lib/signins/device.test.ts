import { describe, expect, it } from 'vitest';
import {
  AttributeDeviceName,
  AttributeIPAddress,
  AttributeUserAgent,
  describeDevice,
  describeUserAgent,
} from './device';

const chromeOnMac =
  'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36';
const safariOnIPhone =
  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1';
const firefoxOnWindows = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0';
const edgeOnWindows =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/126.0.0.0';

describe('describeUserAgent', () => {
  it('names a browser and its system', () => {
    expect(describeUserAgent(chromeOnMac)).toBe('Chrome on macOS');
    expect(describeUserAgent(safariOnIPhone)).toBe('Safari on iOS');
    expect(describeUserAgent(firefoxOnWindows)).toBe('Firefox on Windows');
  });

  it('tells a browser built on another apart from it', () => {
    expect(describeUserAgent(edgeOnWindows)).toBe('Edge on Windows');
  });

  it('shows what it does not recognize as it is', () => {
    expect(describeUserAgent('grpc-node-js/1.12.0')).toBe('grpc-node-js/1.12.0');
  });

  it('has nothing to say about nothing', () => {
    expect(describeUserAgent(undefined)).toBeUndefined();
    expect(describeUserAgent('  ')).toBeUndefined();
  });
});

describe('describeDevice', () => {
  it('prefers the name a client gave its device', () => {
    expect(
      describeDevice({
        [AttributeDeviceName]: 'iPhone 15',
        [AttributeUserAgent]: safariOnIPhone,
        [AttributeIPAddress]: '203.0.113.7',
      }),
    ).toEqual({ name: 'iPhone 15', address: '203.0.113.7' });
  });

  it('reads the browser off the user agent otherwise', () => {
    expect(describeDevice({ [AttributeUserAgent]: chromeOnMac })).toEqual({
      name: 'Chrome on macOS',
      address: undefined,
    });
  });

  it('describes a login nothing was recorded for as nothing', () => {
    expect(describeDevice(undefined)).toEqual({ name: undefined, address: undefined });
    expect(describeDevice({})).toEqual({ name: undefined, address: undefined });
  });
});
