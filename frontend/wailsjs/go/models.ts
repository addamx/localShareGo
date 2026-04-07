export namespace auth {
	
	export class AuthStatus {
	    tokenTtlMinutes: number;
	    rotationEnabled: boolean;
	    bearerHeaderName: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tokenTtlMinutes = source["tokenTtlMinutes"];
	        this.rotationEnabled = source["rotationEnabled"];
	        this.bearerHeaderName = source["bearerHeaderName"];
	    }
	}
	export class LinkedDeviceSummary {
	    id: string;
	    name: string;
	    lastKnownIp: string;
	    lastSeenAt: number;
	    online: boolean;
	    expiresAt: number;
	
	    static createFrom(source: any = {}) {
	        return new LinkedDeviceSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.lastKnownIp = source["lastKnownIp"];
	        this.lastSeenAt = source["lastSeenAt"];
	        this.online = source["online"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class PairRequestSummary {
	    id: string;
	    deviceId: string;
	    deviceName: string;
	    status: string;
	    createdAt: number;
	    updatedAt: number;
	    expiresAt: number;
	
	    static createFrom(source: any = {}) {
	        return new PairRequestSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.deviceId = source["deviceId"];
	        this.deviceName = source["deviceName"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class SessionSnapshot {
	    sessionId: string;
	    sessionKind: string;
	    deviceId: string;
	    deviceName: string;
	    expiresAt: number;
	    status: string;
	    accessUrl: string;
	    publicHost: string;
	    publicPort: number;
	    webBasePath: string;
	    tokenTtlMinutes: number;
	    bearerHeaderName: string;
	    tokenQueryKey: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.sessionKind = source["sessionKind"];
	        this.deviceId = source["deviceId"];
	        this.deviceName = source["deviceName"];
	        this.expiresAt = source["expiresAt"];
	        this.status = source["status"];
	        this.accessUrl = source["accessUrl"];
	        this.publicHost = source["publicHost"];
	        this.publicPort = source["publicPort"];
	        this.webBasePath = source["webBasePath"];
	        this.tokenTtlMinutes = source["tokenTtlMinutes"];
	        this.bearerHeaderName = source["bearerHeaderName"];
	        this.tokenQueryKey = source["tokenQueryKey"];
	    }
	}

}

export namespace clipboard {
	
	export class ClipboardStatus {
	    mode: string;
	    pollIntervalMs: number;
	    dedupWindowMs: number;
	    maxTextBytes: number;
	    currentItemTracking: boolean;
	    running: boolean;
	    subscriberCount: number;
	    refreshEventTopic: string;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.pollIntervalMs = source["pollIntervalMs"];
	        this.dedupWindowMs = source["dedupWindowMs"];
	        this.maxTextBytes = source["maxTextBytes"];
	        this.currentItemTracking = source["currentItemTracking"];
	        this.running = source["running"];
	        this.subscriberCount = source["subscriberCount"];
	        this.refreshEventTopic = source["refreshEventTopic"];
	    }
	}

}

export namespace config {
	
	export class AppPaths {
	    appDir: string;
	    dataDir: string;
	    databasePath: string;
	    fileStagingDir: string;
	    desktopReceiveDir: string;
	    desktopSettingsPath: string;
	    logsDir: string;
	
	    static createFrom(source: any = {}) {
	        return new AppPaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appDir = source["appDir"];
	        this.dataDir = source["dataDir"];
	        this.databasePath = source["databasePath"];
	        this.fileStagingDir = source["fileStagingDir"];
	        this.desktopReceiveDir = source["desktopReceiveDir"];
	        this.desktopSettingsPath = source["desktopSettingsPath"];
	        this.logsDir = source["logsDir"];
	    }
	}
	export class RuntimeConfig {
	    lanHost: string;
	    preferredPort: number;
	    maxTextBytes: number;
	    clipboardPollIntervalMs: number;
	    tokenTtlMinutes: number;
	    databaseFileName: string;
	    webRoute: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lanHost = source["lanHost"];
	        this.preferredPort = source["preferredPort"];
	        this.maxTextBytes = source["maxTextBytes"];
	        this.clipboardPollIntervalMs = source["clipboardPollIntervalMs"];
	        this.tokenTtlMinutes = source["tokenTtlMinutes"];
	        this.databaseFileName = source["databaseFileName"];
	        this.webRoute = source["webRoute"];
	    }
	}

}

export namespace httpserver {
	
	export class HttpServerStatus {
	    bindHost: string;
	    preferredPort: number;
	    effectivePort?: number;
	    state: string;
	    lastError?: string;
	    healthEndpoint: string;
	    webBasePath: string;
	    sseEndpoint: string;
	
	    static createFrom(source: any = {}) {
	        return new HttpServerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bindHost = source["bindHost"];
	        this.preferredPort = source["preferredPort"];
	        this.effectivePort = source["effectivePort"];
	        this.state = source["state"];
	        this.lastError = source["lastError"];
	        this.healthEndpoint = source["healthEndpoint"];
	        this.webBasePath = source["webBasePath"];
	        this.sseEndpoint = source["sseEndpoint"];
	    }
	}
	export class OnlineDevice {
	    id: string;
	    name: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new OnlineDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	    }
	}
	export class SyncClipboardResponse {
	    deliveredDevices: OnlineDevice[];
	    desktopItem?: store.ClipboardItemRecord;
	
	    static createFrom(source: any = {}) {
	        return new SyncClipboardResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deliveredDevices = this.convertValues(source["deliveredDevices"], OnlineDevice);
	        this.desktopItem = this.convertValues(source["desktopItem"], store.ClipboardItemRecord);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace network {
	
	export class NetworkStatus {
	    deviceName: string;
	    accessHost: string;
	    accessHosts: string[];
	    lanDiscoveryEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NetworkStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceName = source["deviceName"];
	        this.accessHost = source["accessHost"];
	        this.accessHosts = source["accessHosts"];
	        this.lanDiscoveryEnabled = source["lanDiscoveryEnabled"];
	    }
	}

}

export namespace runtimeapp {
	
	export class ServiceOverview {
	    clipboard: clipboard.ClipboardStatus;
	    httpServer: httpserver.HttpServerStatus;
	    auth: auth.AuthStatus;
	    session: auth.SessionSnapshot;
	    persistence: store.PersistenceStatus;
	    network: network.NetworkStatus;
	
	    static createFrom(source: any = {}) {
	        return new ServiceOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clipboard = this.convertValues(source["clipboard"], clipboard.ClipboardStatus);
	        this.httpServer = this.convertValues(source["httpServer"], httpserver.HttpServerStatus);
	        this.auth = this.convertValues(source["auth"], auth.AuthStatus);
	        this.session = this.convertValues(source["session"], auth.SessionSnapshot);
	        this.persistence = this.convertValues(source["persistence"], store.PersistenceStatus);
	        this.network = this.convertValues(source["network"], network.NetworkStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RouteOverview {
	    naiveDesktop: string;
	    web: string;
	
	    static createFrom(source: any = {}) {
	        return new RouteOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.naiveDesktop = source["naiveDesktop"];
	        this.web = source["web"];
	    }
	}
	export class AppBootstrap {
	    appName: string;
	    routes: RouteOverview;
	    runtimeConfig: config.RuntimeConfig;
	    paths: config.AppPaths;
	    services: ServiceOverview;
	
	    static createFrom(source: any = {}) {
	        return new AppBootstrap(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.routes = this.convertValues(source["routes"], RouteOverview);
	        this.runtimeConfig = this.convertValues(source["runtimeConfig"], config.RuntimeConfig);
	        this.paths = this.convertValues(source["paths"], config.AppPaths);
	        this.services = this.convertValues(source["services"], ServiceOverview);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectivityCheck {
	    host: string;
	    url: string;
	    tcpOk: boolean;
	    httpOk: boolean;
	    httpStatusLine?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectivityCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.url = source["url"];
	        this.tcpOk = source["tcpOk"];
	        this.httpOk = source["httpOk"];
	        this.httpStatusLine = source["httpStatusLine"];
	        this.error = source["error"];
	    }
	}
	export class ConnectivityReport {
	    bindHost: string;
	    preferredPort: number;
	    effectivePort: number;
	    serverState: string;
	    serverError?: string;
	    accessUrl: string;
	    checks: ConnectivityCheck[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectivityReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bindHost = source["bindHost"];
	        this.preferredPort = source["preferredPort"];
	        this.effectivePort = source["effectivePort"];
	        this.serverState = source["serverState"];
	        this.serverError = source["serverError"];
	        this.accessUrl = source["accessUrl"];
	        this.checks = this.convertValues(source["checks"], ConnectivityCheck);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace settings {
	
	export class DesktopSettings {
	    showAppHotkey: string;
	
	    static createFrom(source: any = {}) {
	        return new DesktopSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.showAppHotkey = source["showAppHotkey"];
	    }
	}

}

export namespace store {
	
	export class ClipboardFileMeta {
	    fileName: string;
	    extension: string;
	    mimeType: string;
	    sizeBytes: number;
	    thumbnailDataUrl?: string;
	    transferState: string;
	    progressPercent: number;
	    localPath?: string;
	    downloadedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardFileMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.extension = source["extension"];
	        this.mimeType = source["mimeType"];
	        this.sizeBytes = source["sizeBytes"];
	        this.thumbnailDataUrl = source["thumbnailDataUrl"];
	        this.transferState = source["transferState"];
	        this.progressPercent = source["progressPercent"];
	        this.localPath = source["localPath"];
	        this.downloadedAt = source["downloadedAt"];
	    }
	}
	export class ClipboardItemRecord {
	    itemKind: string;
	    id: string;
	    content: string;
	    contentType: string;
	    hash: string;
	    preview: string;
	    charCount: number;
	    fileMeta?: ClipboardFileMeta;
	    sourceKind: string;
	    sourceDeviceId?: string;
	    pinned: boolean;
	    isCurrent: boolean;
	    deletedAt?: number;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardItemRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemKind = source["itemKind"];
	        this.id = source["id"];
	        this.content = source["content"];
	        this.contentType = source["contentType"];
	        this.hash = source["hash"];
	        this.preview = source["preview"];
	        this.charCount = source["charCount"];
	        this.fileMeta = this.convertValues(source["fileMeta"], ClipboardFileMeta);
	        this.sourceKind = source["sourceKind"];
	        this.sourceDeviceId = source["sourceDeviceId"];
	        this.pinned = source["pinned"];
	        this.isCurrent = source["isCurrent"];
	        this.deletedAt = source["deletedAt"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ClipboardListQuery {
	    search?: string;
	    pinnedOnly: boolean;
	    includeDeleted: boolean;
	    createdBefore?: number;
	    beforeId?: string;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardListQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.pinnedOnly = source["pinnedOnly"];
	        this.includeDeleted = source["includeDeleted"];
	        this.createdBefore = source["createdBefore"];
	        this.beforeId = source["beforeId"];
	        this.limit = source["limit"];
	    }
	}
	export class PersistenceStatus {
	    databasePath: string;
	    migrationsEnabled: boolean;
	    schemaVersion: number;
	    ready: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PersistenceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.databasePath = source["databasePath"];
	        this.migrationsEnabled = source["migrationsEnabled"];
	        this.schemaVersion = source["schemaVersion"];
	        this.ready = source["ready"];
	    }
	}

}

