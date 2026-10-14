# Primary Test Device: Infinix Note 11S

## Specifications
- **Model**: Infinix Note 11S (X698)
- **OS**: XOS 10.6
- **Android Version**: Android 12 (API level 31)
- **Status**: Primary verified hardware platform for Jerion v1.0

## Verified Device Behaviors & Quirks
- **USB Debugging Activation**:
  1. Navigate to **Settings > My Phone**.
  2. Tap **Build Number** 7 times until the toast notification confirms developer status.
  3. Navigate to **Settings > System > Developer Options**.
  4. Toggle on **USB Debugging**.
  5. Connect to PC via USB cable; when the RSA fingerprint prompt appears on screen, check **"Always allow from this computer"** and select **Allow**.
- **OEM Package Baseline & Protected System Guardrails**:
  The following core Transsion XOS packages are protected in `internal/remediate/guardrail.go` to prevent accidental disabling or uninstallation:
  - `com.transsion.launcher` (XOS Home Screen / Launcher)
  - `com.transsion.systemupdate` (XOS System OTA Updater)
  - `com.transsion.smartpanel` (Transsion Edge Smart Panel)
  - `com.transsion.phonemaster` (Transsion Phone Master Security Center)
- **Background Restrictions & Overlay Permissions**:
  - XOS 10.6 enforces aggressive background restrictions via Power Marathon. Background telemetry for Popup Hunter should be performed with the screen active.
  - Overlay detection inspects `SYSTEM_ALERT_WINDOW` and `TYPE_APPLICATION_OVERLAY` windows via `dumpsys window`.
- **Hardware Quirk Status**:
  - Other vendor-specific settings commands remain marked *unverified* unless directly confirmed on physical device hardware.
