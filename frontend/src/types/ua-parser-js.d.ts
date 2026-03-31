declare module "ua-parser-js" {
  interface BrowserInfo {
    name?: string;
  }

  interface OSInfo {
    name?: string;
  }

  export default class UAParser {
    constructor(userAgent?: string);
    getBrowser(): BrowserInfo;
    getOS(): OSInfo;
  }
}
