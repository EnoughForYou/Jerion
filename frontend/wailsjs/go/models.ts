export namespace adb {
	
	export class DeviceInfo {
	    serial: string;
	    state: string;
	    product?: string;
	    model?: string;
	    device?: string;
	    transportId?: string;
	    rawState: string;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serial = source["serial"];
	        this.state = source["state"];
	        this.product = source["product"];
	        this.model = source["model"];
	        this.device = source["device"];
	        this.transportId = source["transportId"];
	        this.rawState = source["rawState"];
	    }
	}
	export class USBDevice {
	    deviceId: string;
	    name: string;
	    manufacturer: string;
	    pnpClass: string;
	    status: string;
	    configManagerErrorCode: number;
	
	    static createFrom(source: any = {}) {
	        return new USBDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceId = source["deviceId"];
	        this.name = source["name"];
	        this.manufacturer = source["manufacturer"];
	        this.pnpClass = source["pnpClass"];
	        this.status = source["status"];
	        this.configManagerErrorCode = source["configManagerErrorCode"];
	    }
	}
	export class ConnectionDiagnosticResult {
	    adbServerRunning: boolean;
	    adbVersion: string;
	    usbHardwarePresent: boolean;
	    usbDevices: USBDevice[];
	    adbDevices: DeviceInfo[];
	    likelyCause: string;
	    headline: string;
	    possibleCauses: string[];
	    recommendedSteps: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDiagnosticResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.adbServerRunning = source["adbServerRunning"];
	        this.adbVersion = source["adbVersion"];
	        this.usbHardwarePresent = source["usbHardwarePresent"];
	        this.usbDevices = this.convertValues(source["usbDevices"], USBDevice);
	        this.adbDevices = this.convertValues(source["adbDevices"], DeviceInfo);
	        this.likelyCause = source["likelyCause"];
	        this.headline = source["headline"];
	        this.possibleCauses = source["possibleCauses"];
	        this.recommendedSteps = source["recommendedSteps"];
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

export namespace contract {
	
	export class RemediationAction {
	    id: string;
	    type: string;
	    package: string;
	    title: string;
	    impact: string;
	    undoable: boolean;
	    isBlocked: boolean;
	    blockedReason?: string;
	
	    static createFrom(source: any = {}) {
	        return new RemediationAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.package = source["package"];
	        this.title = source["title"];
	        this.impact = source["impact"];
	        this.undoable = source["undoable"];
	        this.isBlocked = source["isBlocked"];
	        this.blockedReason = source["blockedReason"];
	    }
	}
	export class ActionLogEntry {
	    id: string;
	    timestamp: string;
	    action: RemediationAction;
	    status: string;
	    error?: string;
	    backupPath?: string;
	    auditSaved?: boolean;
	    auditError?: string;
	
	    static createFrom(source: any = {}) {
	        return new ActionLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.action = this.convertValues(source["action"], RemediationAction);
	        this.status = source["status"];
	        this.error = source["error"];
	        this.backupPath = source["backupPath"];
	        this.auditSaved = source["auditSaved"];
	        this.auditError = source["auditError"];
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
	export class AllowlistEntry {
	    id: string;
	    type: string;
	    target: string;
	    addedAt: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new AllowlistEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.target = source["target"];
	        this.addedAt = source["addedAt"];
	        this.reason = source["reason"];
	    }
	}
	export class AppInfo {
	    package: string;
	    label: string;
	    versionName: string;
	    versionCode: string;
	    isSystem: boolean;
	    isSideloaded: boolean;
	    installerPackage?: string;
	    targetSdk: number;
	    minSdk: number;
	    installedAt: string;
	    updatedAt: string;
	    permissions: string[];
	    dangerousPermissions: string[];
	    hasOverlayPermission: boolean;
	    hasAccessibilityPermission: boolean;
	    isDeviceAdmin: boolean;
	    isBatteryOptimized: boolean;
	    signatures: string[];
	    suspiciousFlags: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.package = source["package"];
	        this.label = source["label"];
	        this.versionName = source["versionName"];
	        this.versionCode = source["versionCode"];
	        this.isSystem = source["isSystem"];
	        this.isSideloaded = source["isSideloaded"];
	        this.installerPackage = source["installerPackage"];
	        this.targetSdk = source["targetSdk"];
	        this.minSdk = source["minSdk"];
	        this.installedAt = source["installedAt"];
	        this.updatedAt = source["updatedAt"];
	        this.permissions = source["permissions"];
	        this.dangerousPermissions = source["dangerousPermissions"];
	        this.hasOverlayPermission = source["hasOverlayPermission"];
	        this.hasAccessibilityPermission = source["hasAccessibilityPermission"];
	        this.isDeviceAdmin = source["isDeviceAdmin"];
	        this.isBatteryOptimized = source["isBatteryOptimized"];
	        this.signatures = source["signatures"];
	        this.suspiciousFlags = source["suspiciousFlags"];
	    }
	}
	export class AuditItem {
	    id: string;
	    category: string;
	    title: string;
	    status: string;
	    summary: string;
	    evidence: string[];
	    remedy?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.evidence = source["evidence"];
	        this.remedy = source["remedy"];
	    }
	}
	export class Device {
	    id: string;
	    serial: string;
	    model: string;
	    manufacturer: string;
	    androidVersion: string;
	    sdkInt: number;
	    batteryLevel: number;
	    isCharging: boolean;
	    connectionState: string;
	    isEmulator: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.serial = source["serial"];
	        this.model = source["model"];
	        this.manufacturer = source["manufacturer"];
	        this.androidVersion = source["androidVersion"];
	        this.sdkInt = source["sdkInt"];
	        this.batteryLevel = source["batteryLevel"];
	        this.isCharging = source["isCharging"];
	        this.connectionState = source["connectionState"];
	        this.isEmulator = source["isEmulator"];
	    }
	}
	export class Evidence {
	    type: string;
	    title: string;
	    details: string;
	    rawOutput?: string;
	
	    static createFrom(source: any = {}) {
	        return new Evidence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.title = source["title"];
	        this.details = source["details"];
	        this.rawOutput = source["rawOutput"];
	    }
	}
	export class Finding {
	    id: string;
	    ruleId: string;
	    title: string;
	    severity: string;
	    confidence: string;
	    package?: string;
	    whyFlagged: string;
	    falsePositiveHint: string;
	    evidence: Evidence[];
	    recommendedAction?: RemediationAction;
	
	    static createFrom(source: any = {}) {
	        return new Finding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ruleId = source["ruleId"];
	        this.title = source["title"];
	        this.severity = source["severity"];
	        this.confidence = source["confidence"];
	        this.package = source["package"];
	        this.whyFlagged = source["whyFlagged"];
	        this.falsePositiveHint = source["falsePositiveHint"];
	        this.evidence = this.convertValues(source["evidence"], Evidence);
	        this.recommendedAction = this.convertValues(source["recommendedAction"], RemediationAction);
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
	export class InventoryFilter {
	    category: string;
	    search: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.search = source["search"];
	    }
	}
	export class MonitorEvent {
	    id: string;
	    timestamp: string;
	    type: string;
	    package: string;
	    details: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new MonitorEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.type = source["type"];
	        this.package = source["package"];
	        this.details = source["details"];
	    }
	}
	export class MonitorCandidate {
	    package: string;
	    appLabel: string;
	    confidence: string;
	    score: number;
	    timelineEvents: MonitorEvent[];
	    whyFlagged: string;
	
	    static createFrom(source: any = {}) {
	        return new MonitorCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.package = source["package"];
	        this.appLabel = source["appLabel"];
	        this.confidence = source["confidence"];
	        this.score = source["score"];
	        this.timelineEvents = this.convertValues(source["timelineEvents"], MonitorEvent);
	        this.whyFlagged = source["whyFlagged"];
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
	
	export class MonitorStatus {
	    isActive: boolean;
	    startedAt?: string;
	    sourcesActive: string[];
	
	    static createFrom(source: any = {}) {
	        return new MonitorStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActive = source["isActive"];
	        this.startedAt = source["startedAt"];
	        this.sourcesActive = source["sourcesActive"];
	    }
	}
	export class MonitorSnapshot {
	    status: MonitorStatus;
	    events: MonitorEvent[];
	    candidates: MonitorCandidate[];
	
	    static createFrom(source: any = {}) {
	        return new MonitorSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = this.convertValues(source["status"], MonitorStatus);
	        this.events = this.convertValues(source["events"], MonitorEvent);
	        this.candidates = this.convertValues(source["candidates"], MonitorCandidate);
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
	
	export class PreflightCheck {
	    id: string;
	    title: string;
	    status: string;
	    detail: string;
	    remedy?: string;
	    sourceUrl?: string;
	    downloadProgress?: number;
	
	    static createFrom(source: any = {}) {
	        return new PreflightCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.detail = source["detail"];
	        this.remedy = source["remedy"];
	        this.sourceUrl = source["sourceUrl"];
	        this.downloadProgress = source["downloadProgress"];
	    }
	}
	
	export class RemediationPlan {
	    actions: RemediationAction[];
	    requiresBackup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RemediationPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.actions = this.convertValues(source["actions"], RemediationAction);
	        this.requiresBackup = source["requiresBackup"];
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
	export class ReportExportOptions {
	    scanId: string;
	    format: string;
	    targetPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportExportOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scanId = source["scanId"];
	        this.format = source["format"];
	        this.targetPath = source["targetPath"];
	    }
	}
	export class ScanSummaryItem {
	    id: string;
	    startedAt: string;
	    completedAt?: string;
	    mode: string;
	    verdict: string;
	    findingsCount: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new ScanSummaryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.startedAt = source["startedAt"];
	        this.completedAt = source["completedAt"];
	        this.mode = source["mode"];
	        this.verdict = source["verdict"];
	        this.findingsCount = source["findingsCount"];
	    }
	}
	export class ScanComparison {
	    scanA: ScanSummaryItem;
	    scanB: ScanSummaryItem;
	    newFindings: Finding[];
	    resolvedFindings: Finding[];
	    unchangedFindings: Finding[];
	
	    static createFrom(source: any = {}) {
	        return new ScanComparison(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scanA = this.convertValues(source["scanA"], ScanSummaryItem);
	        this.scanB = this.convertValues(source["scanB"], ScanSummaryItem);
	        this.newFindings = this.convertValues(source["newFindings"], Finding);
	        this.resolvedFindings = this.convertValues(source["resolvedFindings"], Finding);
	        this.unchangedFindings = this.convertValues(source["unchangedFindings"], Finding);
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
	export class ScanOptions {
	    mode: string;
	    includeSystem: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ScanOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.includeSystem = source["includeSystem"];
	    }
	}
	export class SecuritySummary {
	    counts: Record<string, number>;
	    totalApps: number;
	    lastScanAt?: string;
	    verdict: string;
	
	    static createFrom(source: any = {}) {
	        return new SecuritySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.counts = source["counts"];
	        this.totalApps = source["totalApps"];
	        this.lastScanAt = source["lastScanAt"];
	        this.verdict = source["verdict"];
	    }
	}
	export class ScanResult {
	    id: string;
	    deviceId: string;
	    mode: string;
	    startedAt: string;
	    completedAt?: string;
	    status: string;
	    findings: Finding[];
	    summary: SecuritySummary;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.deviceId = source["deviceId"];
	        this.mode = source["mode"];
	        this.startedAt = source["startedAt"];
	        this.completedAt = source["completedAt"];
	        this.status = source["status"];
	        this.findings = this.convertValues(source["findings"], Finding);
	        this.summary = this.convertValues(source["summary"], SecuritySummary);
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
	
	
	export class Settings {
	    theme: string;
	    adbPath?: string;
	    customAdbPort?: number;
	    deepScanTimeoutSec: number;
	    autoCheckPreflight: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.adbPath = source["adbPath"];
	        this.customAdbPort = source["customAdbPort"];
	        this.deepScanTimeoutSec = source["deepScanTimeoutSec"];
	        this.autoCheckPreflight = source["autoCheckPreflight"];
	    }
	}

}

export namespace main {
	
	export class AppSystemInfo {
	    appVersion: string;
	    rulesVersion: string;
	    rulesCount: number;
	    adbVersion?: string;
	    contractVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appVersion = source["appVersion"];
	        this.rulesVersion = source["rulesVersion"];
	        this.rulesCount = source["rulesCount"];
	        this.adbVersion = source["adbVersion"];
	        this.contractVersion = source["contractVersion"];
	    }
	}
	export class ClamAVConfig {
	    enabled: boolean;
	    host: string;
	
	    static createFrom(source: any = {}) {
	        return new ClamAVConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.host = source["host"];
	    }
	}
	export class ClamAVStatus {
	    connected: boolean;
	    version?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ClamAVStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.version = source["version"];
	        this.error = source["error"];
	    }
	}

}

