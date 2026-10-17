export type BridgeErrorCode =
  | 'AdbMissing'
  | 'DeviceUnauthorized'
  | 'DeviceOffline'
  | 'Timeout'
  | 'ParseFailure'
  | 'PermissionDenied'
  | 'NotSupportedOnThisOS'
  | 'ActionBlocked'
  | 'ValidationFailed'
  | 'CommandFailed'
  | 'InternalError';

export interface BridgeErrorPayload {
  code: BridgeErrorCode;
  message: string;
  remedy?: string;
}

export const WIRE_ERROR_PREFIX = 'JERION_ERR:';

export class BridgeError extends Error {
  public readonly code: BridgeErrorCode;
  public readonly remedy?: string;

  constructor(code: BridgeErrorCode, message: string, remedy?: string) {
    super(message);
    this.name = 'BridgeError';
    this.code = code;
    this.remedy = remedy;
    Object.setPrototypeOf(this, BridgeError.prototype);
  }

  toWire(): string {
    const payload: BridgeErrorPayload = {
      code: this.code,
      message: this.message,
      ...(this.remedy ? { remedy: this.remedy } : {}),
    };
    return `${WIRE_ERROR_PREFIX}${JSON.stringify(payload)}`;
  }

  static fromWire(raw: unknown): BridgeError {
    if (raw instanceof BridgeError) {
      return raw;
    }

    const rawString =
      typeof raw === 'string'
        ? raw
        : raw instanceof Error
          ? raw.message
          : typeof raw === 'object' && raw !== null && 'message' in raw && typeof (raw as { message: unknown }).message === 'string'
            ? (raw as { message: string }).message
            : String(raw);

    if (rawString.startsWith(WIRE_ERROR_PREFIX)) {
      const jsonStr = rawString.slice(WIRE_ERROR_PREFIX.length);
      try {
        const parsed = JSON.parse(jsonStr) as Partial<BridgeErrorPayload>;
        if (parsed && typeof parsed.code === 'string' && typeof parsed.message === 'string') {
          return new BridgeError(parsed.code as BridgeErrorCode, parsed.message, parsed.remedy);
        }
      } catch {
        // Malformed wire JSON falls through to InternalError
      }
    }

    return new BridgeError('InternalError', rawString);
  }
}
