/**
 * What the API recorded about the device behind a login, as ListSignIns' attributes carry it:
 * the address and user agent it was last renewed from, and the name a client gave its device.
 * See internal/authentication/devices in the backend.
 */
export const AttributeIPAddress = 'ip_address';
export const AttributeUserAgent = 'user_agent';
export const AttributeDeviceName = 'device_name';

/** SignInDevice is a login's device, described for a person rather than for a log. */
export interface SignInDevice {
  /** name is the device's own name, or a browser and system read off its user agent. */
  name?: string;
  /** address is where the login was last renewed from. */
  address?: string;
}

const browsers: [RegExp, string][] = [
  [/Edg(A|iOS)?\//, 'Edge'],
  [/OPR\//, 'Opera'],
  [/Firefox\/|FxiOS\//, 'Firefox'],
  [/Chrome\/|CriOS\//, 'Chrome'],
  [/Safari\//, 'Safari'],
];

const systems: [RegExp, string][] = [
  [/iPhone|iPad|iPod/, 'iOS'],
  [/Android/, 'Android'],
  [/Mac OS X|Macintosh/, 'macOS'],
  [/Windows/, 'Windows'],
  [/CrOS/, 'ChromeOS'],
  [/Linux/, 'Linux'],
];

/**
 * describeUserAgent is "Browser on System" for a browser it recognizes, and the user agent itself
 * for anything else — a script, an app, a client nobody planned for — since that is still more
 * than nothing to somebody deciding whether a login is theirs.
 */
export function describeUserAgent(userAgent: string | undefined): string | undefined {
  const ua = userAgent?.trim();
  if (!ua) {
    return undefined;
  }

  const browser = browsers.find(([pattern]) => pattern.test(ua))?.[1];
  const system = systems.find(([pattern]) => pattern.test(ua))?.[1];

  if (browser && system) {
    return `${browser} on ${system}`;
  }
  return browser ?? system ?? ua;
}

/** describeDevice reads a listed login's attributes. */
export function describeDevice(attributes: Record<string, string> | undefined): SignInDevice {
  const deviceName = attributes?.[AttributeDeviceName]?.trim();
  const address = attributes?.[AttributeIPAddress]?.trim();

  return {
    name: deviceName || describeUserAgent(attributes?.[AttributeUserAgent]),
    address: address || undefined,
  };
}
