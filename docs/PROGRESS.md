# PROGRESS

## Status sekarang
- Fase aktif: HONESTY-1.1 — Version & Contract Single Source of Truth, Content-Hash Rule Version, ClamAV Deduplication (v1.2.1)
- Branch: phase/safety-1
- Terakhir selesai: HONESTY-1 — Product Truthfulness, Fail-Closed ClamAV & Dynamic Telemetry (tag v1.2.0)

## Fase HONESTY-1.1 (Version & Contract Single Source, Content-Hash Rules, ClamAV Dedup) — checklist
- [x] Sub-tugas H-1.1.A: Sinkronisasi versi tunggal `1.2.1` pada `internal/version/version.go`, `wails.json`, `frontend/package.json`, `frontend/package-lock.json`, dan `CHANGELOG.md` dengan pengujian konsistensi otomatis `TestVersionConsistency` di `internal/version/version_test.go`.
- [x] Sub-tugas H-1.1.B: Single source of truth versi kontrak `1.1.0` via konstanta `version.ContractVersion` di Go backend, eliminasi sufiks manual `(Additive)` pada UI (`SettingsView.tsx`), pembaruan schema Zod/TS di `schema.ts`, `index.ts`, dan dokumentasi `docs/contract.md`.
- [x] Sub-tugas H-1.1.C: Penurunan versi aturan forensik secara dinamis dari SHA-256 digest katalog YAML embedded (`2026.03.1+<8hex>`), eliminasi konstanta hardcoded `RULE_VERSION` pada frontend (`deviceSecurityState.ts`, `StatusBar.tsx`, dll.), penyediaan `useSystemInfoStore` Zustand untuk penyebaran metadata dinamis.
- [x] Sub-tugas H-1.1.D: Deduplikasi error daemon ClamAV pada Deep Scan menjadi tepat satu temuan tingkat scan (`CLAMAV-SCAN-ERR`), pencegahan perulangan panggilan ke daemon yang mati, penandaan scan parsial jujur, dan penambahan catatan cakupan eksplisit ("Used during Deep Scan only (base APK inspection)") pada i18n & UI.

## Pengukuran & Kualitas (HONESTY-1.1)
- **Backend Go Test Suite**: 100% test lulus di seluruh 17 paket backend Go (`go test -count=1 . ./internal/...`).
- **Backend Go Vet**: Lulus bersih tanpa peringatan (`go vet . ./internal/...`).
- **Frontend Vitest Suite**: 159 pengujian lolos di 23 file tes (`npm run test`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock di seluruh 3 bundle produksi.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 22.04 detik.

## Fase HONESTY-1 (Product Truthfulness & Dynamic Verification) — riwayat selesai (v1.2.0)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`go test -count=1 . ./internal/...`).
- **Backend Go Vet**: Lulus bersih tanpa peringatan (`go vet . ./internal/...`).
- **Frontend Vitest Suite**: 159 pengujian lolos di 23 file tes (`npm run test`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock di seluruh bundle produksi.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 22.65 detik.

## Fase SEC-1 (ClamAV Hardening & Stream Limit Handling) — riwayat selesai (v1.1.1)

## Fase UX-1b (Visual Polish & System Theme) — riwayat selesai (v1.1.0)
- Mesin 8 status keamanan kanonis (`no-device`, `unauthorized`, `offline`, `never-scanned`, `scanning`, `scanned-clean`, `scanned-findings`, `partial-failed`) dengan pengikatan serial perangkat ketat.
- Eliminasi rasa aman palsu: centang hijau eksklusif hanya untuk `scanned-clean`.
- Single source of truth bridge via `bridgeKind`, badge MOCK dikunci strictly pada `import.meta.env.DEV`.
- Prioritas status perangkat sebelum error state.
- 152 unit tests lulus, 0 marker mock di bundle produksi.

## Fase HYG-1 (Kebersihan Repo & CI) — riwayat selesai (v1.0.1)
- **Produksi NPM Audit**: 0 vulnerabilities (`npm audit --omit=dev --prefix frontend`).
- **Audit Tooling Dev Penuh (Accepted Risk)**: 7 temuan (braces DoS via nested patterns, postcss CPU exhaustion) terbatas pada AST parser watcher Tailwind v3 lokal dan tidak terbundel ke runtime JavaScript browser maupun binary executable Wails.
- **Backend Go Vet**: Lulus bersih tanpa peringatan pada `. ./internal/...`.
- **Backend Go Test Suite**: 100% lulus pada `. ./internal/...` tanpa gangguan modul pihak ketiga di node_modules.

## Fase Audit Hardening & Cross-Platform CI Fix — riwayat selesai (v1.0.0)
- [x] Sub-tugas FIX.1: Portabilitas tes checksum platform-tools lintas OS (`internal/preflight/preflight_test.go`) dengan memvalidasi hash default menggunakan `runtime.GOOS` dan pengujian tabel mandiri untuk katalog Windows, Linux, dan Darwin (macOS), menyelesaikan kegagalan runner CI Ubuntu GitHub Actions (`TestTrustedPlatformToolsHashes`, `TestStrictChecksumMatching`)
- [x] Sub-tugas FIX.2: Penerusan error status otorisasi perangkat (`DeviceUnauthorized`) pada `RemediateCreatePlan` dan `RemediateExecute`, mencegah penyamaran error dialog RSA USB debugging menjadi `DeviceOffline` sehingga pengguna tidak menerima pesan keliru saat belum menekan 'Allow'
- [x] Sub-tugas FIX.3: Klasifikasi akurat aplikasi sistem yang diperbarui (updated system apps dengan APK di `/data/app/`) pada remediasi via metadata Package Manager tepercaya (`pm list packages -s`, flags dumpsys `SYSTEM`/`UPDATED_SYSTEM_APP`, dan fallback mount point) di `internal/inventory/parser.go`, `internal/inventory/inventory.go`, dan `app.go`
- [x] Sub-tugas FIX.4: Penanganan seluruh status terminal scan (`completed`, `partial`, `failed`, `cancelled`) pada polling sinkron `ScanQuick()`, mengeliminasi blocking 60 detik saat scan berakhir dengan status terminal non-completed
- [x] Sub-tugas FIX.5: Unit test komprehensif table-driven pada `app_test.go` (`TestRemediateDeviceUnauthorizedForwarding`, `TestRemediateUpdatedSystemApp`, `TestScanQuickTerminalStates`), `inventory_test.go` (`TestParseDumpsysPackage_UpdatedSystemApp`, `TestIsSystemPackage`), dan `preflight_test.go`
- [x] Sub-tugas FIX.6: Resolusi pelanggaran arsitektur layering pada `app.go` (`isSystemPackage` mendelegasikan langsung ke `internal/inventory.Service`), pencocokan token flag presisi pada `ParseDumpsysPackage` untuk mengeliminasi false positive `REQUIRED_FOR_SYSTEM_USER`, sanitasi nama paket via `ValidatePackageName`, standardisasi path prefix partisi sistem kanonis (`isSystemMountPath`), dan penambahan pengujian tepi komprehensif

## Pengukuran & Kualitas (Audit Hardening & Cross-Platform CI Fix)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`go test -count=1 ./...`).
- **Go Vet**: Lulus bersih tanpa peringatan (`go vet ./...`).
- **Frontend Vitest Suite**: 136 pengujian lolos di 21 file tes (`npm run test`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock di seluruh bundle produksi.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 18.47 detik.

## Fase Device Detection & Screen Resilience Hardening — riwayat selesai

## Fase Runner ADB Path Sync & Mock Audit — riwayat selesai
- [x] Sub-tugas SYNC.1: Sinkronisasi aktif path biner ADB dari `adb.Manager` ke `adb.Runner` (`RealRunner.SetADBPath`) pada `ResolveADBPath()`, `SetCustomPath()`, `CheckVersion()`, dan `StartServer()`, mengeliminasi kegagalan `exec: "adb": executable file not found in %PATH%`
- [x] Sub-tugas SYNC.2: Resolusi defensif mandiri di `RealRunner.Run()` yang secara otomatis mengadopsi direktori default tools aplikasi (`%LOCALAPPDATA%\Jerion\tools\platform-tools\adb.exe`) saat biner tidak ditemukan pada system PATH
- [x] Sub-tugas SYNC.3: Sinkronisasi runner saat startup aplikasi di `NewAppWithDeps` dan saat pembaruan pengaturan di `SettingsUpdate`
- [x] Sub-tugas SYNC.4: Audit menyeluruh mock data di seluruh proyek — konfirmasi 0 mock data di bundle produksi (`check:no-mock`), 0 mock objects di backend produksi Go, dan data real-time penuh dari ADB hardware & SQLite
- [x] Sub-tugas SYNC.5: Penambahan isolasi environment test pada `preflight_test.go` dan unit test verifikasi sinkronisasi runner path pada `adb_test.go`

## Pengukuran & Kualitas (Runner ADB Path Sync & Mock Audit)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`go test -count=1 ./...`).
- **Go Vet**: Lulus bersih tanpa peringatan (`go vet ./...`).
- **Frontend Vitest Suite**: 127 pengujian lolos di 20 file tes (`npm run test`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock di seluruh bundle produksi.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 18.83 detik.
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`internal/logx`, `internal/secrets`, `internal/store`, `internal/adb`, `internal/preflight`, `internal/inventory`, `internal/apk`, `internal/rules`, `internal/engine`, `internal/intel`, `internal/audit`, `internal/monitor`, `internal/remediate`, `internal/report`, `internal/contract`, dan root `main`) dengan `go test -count=1 ./...`.
- **Go Vet**: Lulus bersih tanpa peringatan (`go vet ./...`).
- **Strict DeviceID Binding**: Cache scan result tanpa DeviceID atau dengan serial berbeda tertolak langsung di `RemediateCreatePlan`, dengan nil-guard aman pada cache iteration.
- **Atomic Batch Reservation**: Mutex eksklusif mencegah eksekusi duplikat atau concurrency race pada seluruh action ID dalam satu batch plan; slice hasil diinisialisasi non-nil mencegah null serialization.
- **Audit Persistence Fidelity**: Status `auditSaved` dan `auditError` dilaporkan secara eksplisit bahkan saat database store unconfigured/nil; UI `RemediateView.tsx` menutup modal dan menampilkan banner error saat eksekusi ditolak backend.
- **Partial Scan Honesty**: Scan parsial dengan 0 temuan menghasilkan verdict "review" baik saat live maupun saat dimuat dari database store (`convertStoreScanRecord`); UI memprioritaskan peringatan parsial dan terminal scan di atas status filter kosong.
- **AppOps Parser Accuracy**: Parser `ParseAppOpsOverlay` dan `ParseSingleAppOpOverlay` membedakan output recognized dan unparseable dengan case-insensitivity dan pemisahan token UID/brackets; `InventoryGet` menerapkan fallback tangguh dari `cmd appops` ke `appops`.
- **Frontend Vitest Suite**: 125 pengujian lolos di 20 file tes (`npm run test`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 17.82 detik.

## Fase Fidelity Hardening Round 4 (Fail-Closed Device Binding, Pre-validation, Scan Verdict Honesty & Downloader Regression Test) — riwayat selesai
- [x] Sub-tugas H4.1: Pengikatan rencana remediasi fail-closed ke serial perangkat (`wireError("DeviceOffline")` saat perangkat kosong/offline di `RemediateCreatePlan` dan `RemediateExecute`), serta penyaringan finding store SQLite hanya untuk serial perangkat aktif
- [x] Sub-tugas H4.2: Pra-validasi (pre-validation) seluruh action ID dan kecocokan serial sebelum mengeksekusi aksi pertama di perangkat, serta rekonstruksi hasil DTO kanonis dari cache backend terpercaya
- [x] Sub-tugas H4.3: Kejujuran verdict scan (`Summary.Verdict = "review"` pada scan failed/cancelled) dan penyembunyian verdict semu pada UI frontend dengan alert notice yang jelas
- [x] Sub-tugas H4.4: Ketepatan runtime three-state (flag `overlay_status_unknown` saat AppOps gagal, fallback `dpm list-owners` untuk Device Admin dengan flag `device_admin_unknown`, dan penanganan whitelist baterai kosong/gagal dengan `battery_status_unknown`)
- [x] Sub-tugas H4.5: Pelaporan eksplisit status persistensi audit log (`auditSaved: *bool`, `auditError: *string`) pada `ActionLogEntry` di backend DTO, TS bridge schema, dan Wails models tanpa membatalkan aksi perangkat yang telah sukses
- [x] Sub-tugas H4.6: Pengujian regresi end-to-end `httptest.Server` pada `DownloadAndExtract` yang memverifikasi penolakan langsung checksum SHA-256 yang tidak cocok dan penerimaan arsip sah

## Fase Fidelity Hardening Round 3 (Device Binding, Remediation Failure Reporting, AppOps & Runtime Truth) — riwayat selesai
- [x] Sub-tugas H3.1: Pengikatan rencana remediasi ke serial perangkat (`DeviceSerial`), penolakan eksekusi lintas perangkat (`DeviceMismatch`), invalidasi aksi single-use dari cache backend, dan pembersihan cache saat perangkat berganti/terputus
- [x] Sub-tugas H3.2: Pelaporan terstruktur kegagalan remediasi (`Status: "failed"`, `Error: &errStr`) pada guardrail/backup failure di `remediate.Execute` dan `RemediateExecute`, tersimpan ke SQLite audit log tanpa kehilangan catatan
- [x] Sub-tugas H3.3: Verifikasi status aktif izin overlay via AppOps (`SYSTEM_ALERT_WINDOW` mode allow) via `dumpsys appops` dan `appops get`, mengeliminasi kesalahan pembacaan dari install permissions
- [x] Sub-tugas H3.4: Pendeteksian menyeluruh active Device Admin via `dumpsys device_policy` ("Enabled Device Admins") dan regex komponen lengkap (`(?:admin=)?([a-zA-Z0-9_.]+)/`)
- [x] Sub-tugas H3.5: Penanganan status optimasi baterai secara jujur (unknown saat pembacaan whitelist gagal) tanpa kesimpulan semu aplikasi dioptimasi
- [x] Sub-tugas H3.6: Unifikasi konversi tipe evidence kanonis (`mapEvidenceType`) pada scan langsung (`convertEngineScanResult`) dan scan persisten (`convertStoreFinding`)
- [x] Sub-tugas H3.7: Emisi terminal failure event (`scan:done` dengan status failed/cancelled) saat scan gagal atau dibatalkan untuk mencegah UI macet
- [x] Sub-tugas H3.8: Penegakan verifikasi checksum ketat (`expectedSHA`) pada unduhan resmi Google Platform-Tools r37.0.1 tanpa fallback ke hash versi lama

## Fase Fidelity Hardening Round 2 (Checksum, Remediation Safety, Runtime Fidelity & Persistence) — riwayat selesai
- [x] Sub-tugas H2.1: Penguncian URL arsip unduhan ADB ke Google Platform-Tools r37.0.1 resmi (`platform-tools_r37.0.1-win.zip`) dan pembaruan katalog SHA-256 (`45f4d631...` Windows, `ee39ad59...` Darwin, `d230f138...` Linux)
- [x] Sub-tugas H2.2: Ekstraksi permission bersih tanpa prefiks (`Declared/Requested:`) pada `RemediateCreatePlan` dan penolakan aksi target kosong pada `internal/remediate`
- [x] Sub-tugas H2.3: Validasi ketat rencana aksi backend pada `RemediateExecute` dan penolakan action ID yang tidak dikenal (`InvalidAction`)
- [x] Sub-tugas H2.4: Akurasi metadata runtime: overlay aktif dari `GrantedPermissions`, status live Accessibility & Device Admin, optimasi baterai dari `dumpsys deviceidle whitelist`, dan mitigasi visual baterai `--`
- [x] Sub-tugas H2.5: Persistensi statistik scan (`TotalApps`) di SQLite schema v2 (`scan_results.total_apps`), pelaporan eksplisit error `SaveScanResult`, dan scoping unik finding ID per scan
- [x] Sub-tugas H2.6: Penegakan strict YAML decoding dengan `KnownFields(true)` pada rule engine loader

## Fase Fidelity Hardening (P1: Evidence Preservation & Canonical Scan ID) — riwayat selesai
- [x] Sub-tugas P1.1: Penyelarasan ID Kanonis Scan (`ScanOptions.ScanID`) antara `app.go:ScanStart` dan `internal/engine/scanner.go:Scan`
- [x] Sub-tugas P1.2: Pelaporan status scan parsial nyata (`Partial: isPartial`) bila ekstraksi/inspeksi APK pada Deep Scan mengalami kegagalan
- [x] Sub-tugas P1.3: Pemetaan delta temuan forensik lengkap pada `ReportCompare()` (`NewFindings`, `ResolvedFindings`, `UnchangedFindings`) via `convertStoreFinding`
- [x] Sub-tugas P1.4: Perhitungan agregasi 5 tingkat severity & penentuan verdict dinamis pada `store.ListScanSummaries` dan `app.go:ReportList`
- [x] Sub-tugas P1.5: Pemanfaatan metadata perangkat nyata dari SQLite pada `ReportExport` (HTML & JSON) tanpa mock

## Fase Fidelity Hardening (P0: Data Fidelity & Remediation Safety) — riwayat selesai
- [x] Sub-tugas P0.1: Eliminasi fallback data perangkat semu (`Android 12`, `SDK 31`, `85%`) pada `convertDevice` di `app.go`
- [x] Sub-tugas P0.2: Penegakan syarat wajib backup APK sebelum uninstall aplikasi pengguna pada `internal/remediate` (`BackupFailed` guardrail)
- [x] Sub-tugas P0.3: Pengkabelan rencana remediasi nyata berdasarkan finding ID tanpa dummy `com.adware.sample` (`RemediateCreatePlan` & `RemediateExecute`)
- [x] Sub-tugas P0.4: Verifikasi integritas checksum SHA-256 wajib pada unduhan ADB platform-tools Google (ADR-0002)
- [x] Sub-tugas P0.5: Akurasi metadata aplikasi (`targetSdk`, `minSdk`, timestamp kosong jika tidak ada) & ringkasan scan `TotalApps` dinamis

## Fase 6 (Tahap T6) — riwayat selesai
- [x] Sub-tugas 6.1: Konfigurasi Windows NSIS Installer & Packaging Final
- [x] Sub-tugas 6.2: Audit Keamanan & Guardrail Produksi (PRD §6.8)
- [x] Sub-tugas 6.3: Dokumentasi Final & Panduan Perangkat Pengguna
- [x] Sub-tugas 6.4: Gerbang Selesai-Tahap T6 (Rilis v1.0.0)

## Pengukuran & Kualitas (Tahap T6 / v1.0.0 Production Release)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`internal/logx`, `internal/secrets`, `internal/store`, `internal/adb`, `internal/preflight`, `internal/inventory`, `internal/apk`, `internal/rules`, `internal/engine`, `internal/intel`, `internal/audit`, `internal/monitor`, `internal/remediate`, `internal/report`, `internal/contract`, dan root `main`).
- **Input Sanitization & Injection Defense**: 100% lulus pada seluruh uji injeksi shell (pipes, backticks, subshells, redirections, newlines) pada nama paket dan nama izin Android.
- **Frontend Vitest Suite**: 116 pengujian lolos di 20 file tes.
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock atau skema runtime terdeteksi pada bundle produksi Vite.
- **Production Compilation (`wails build`)**: Standalone binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 18.4 detik dengan metadata Windows v1.0.0.
- **Security & Privacy Posture (PRD §6.8)**: Keyring OS native tanpa kebocoran credential ke UI, zero remote CDN dependencies, zero telemetry, proteksi denylist sistem kritis (Android Core + Transsion XOS 10.6), dan backup APK sebelum remediate.

## Fase 5 (Tahap T5) — riwayat selesai
- [x] Sub-tugas 5.1: Bridge Runtime Auto-Detection & Bump Versi Kontrak `1.0.0`
- [x] Sub-tugas 5.2: Hardening Ketahanan ADB & Kasus Tepi (Edge Cases)
- [x] Sub-tugas 5.3: Perluasan Rules Forensik & Baseline XOS 10.6 (Infinix Note 11S)
- [x] Sub-tugas 5.4: UI Polish, Aksesibilitas (WCAG AA) & Checklist Anti-AI-Slop
- [x] Sub-tugas 5.5: Konfigurasi Packaging & Windows Installer
- [x] Sub-tugas 5.6: Gerbang Selesai-Tahap T5 (tag v0.5.0)

## Pengukuran & Kualitas (Tahap T5)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`internal/logx`, `internal/secrets`, `internal/store`, `internal/adb`, `internal/preflight`, `internal/inventory`, `internal/apk`, `internal/rules`, `internal/engine`, `internal/intel`, `internal/audit`, `internal/monitor`, `internal/remediate`, `internal/report`, `internal/contract`, dan root `main`).
- **Contract Fidelity (Strict Two-Way)**: Bridge contract `1.0.0 (FROZEN)` terverifikasi dua arah dengan auto-detection runtime Wails (`window.go?.main?.App`).
- **Frontend Vitest Suite**: 116 pengujian lolos di 20 file tes (termasuk pengujian bridge runtime auto-detection).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock atau skema runtime terdeteksi pada bundle produksi Vite.
- **Production Compilation (`wails build`)**: Binary Windows x64 `build/bin/Jerion.exe` terkompilasi bersih dalam 17.6 detik dengan metadata Windows v0.5.0.
- **Accessibility & Anti-AI-Slop**: Kepatuhan penuh WCAG 2.1 AA Focus Visible (`:focus-visible` ring), token-based scrollbar & text selection, `prefers-reduced-motion` override, serta zero-slop design system.

## Fase 4 (Tahap T4) — riwayat selesai
- [x] Sub-tugas 4.1: Paket fondasi & penyimpanan (`internal/logx`, `internal/secrets`, `internal/store` via pure-Go SQLite `modernc.org/sqlite`)
- [x] Sub-tugas 4.2: Lapisan ADB Core & Preflight (`internal/adb`, `internal/preflight`)
- [x] Sub-tugas 4.3: Inventaris Aplikasi & Analisis APK (`internal/inventory`, `internal/apk`)
- [x] Sub-tugas 4.4: Scan Engine, Rules & Intel (`internal/engine`, `internal/rules`, `internal/intel`)
- [x] Sub-tugas 4.5: System Audit & Popup Hunter Monitor (`internal/audit`, `internal/monitor`)
- [x] Sub-tugas 4.6: Remediasi & Laporan (`internal/remediate`, `internal/report`)
- [x] Sub-tugas 4.7: Service Layer & Go DTOs (`app.go` Wails bindings)
- [x] Sub-tugas 4.8: Gerbang selesai-tahap T4 (tag v0.4.0)

## Pengukuran & Kualitas (Tahap T4)
- **Backend Go Test Suite**: 100% test lulus di seluruh 16 paket backend Go (`internal/logx`, `internal/secrets`, `internal/store`, `internal/adb`, `internal/preflight`, `internal/inventory`, `internal/apk`, `internal/rules`, `internal/engine`, `internal/intel`, `internal/audit`, `internal/monitor`, `internal/remediate`, `internal/report`, `internal/contract`, dan root `main`).
- **Contract Fidelity (Strict Two-Way)**: 21 dari 21 fixture kontrak JSON divalidasi dengan `json.NewDecoder.DisallowUnknownFields()` dan round-trip marshaling di `internal/contract/contract_test.go` tanpa ada field yang hilang atau tidak dikenali.
- **Frontend Vitest Suite**: 114 pengujian lolos di 19 file tes (`src/features/monitor/monitor.test.tsx`, `src/features/audit/audit.test.tsx`, `src/features/apps/apps.test.tsx`, `src/features/settings/settings.test.tsx`, `src/features/scan/scan.test.tsx`, dll.).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker mock atau skema runtime terdeteksi pada bundle produksi Vite.
- **Production Compilation (`wails build`)**: Binary Windows x64 `build/bin/Jerion.exe` berhasil dikompilasi secara bersih dalam 21.8 detik.
- **Safety Guardrails & Anti-AI-Slop**: 100% dipatuhi (Windows `HideWindow: true`, regex sanitasi nama paket `^[a-zA-Z0-9_.]+$`, denylist sistem kritis dilindungi, SQLite WAL mode tanpa CGO, offline report zero-CDN, tanpa skor persen sembarangan).

## Fase 3 (Tahap T3) — riwayat selesai
- [x] Sub-tugas 3.1: Interactive Home view & OnboardingModal
- [x] Sub-tugas 3.2: Interactive Scan view with timeline, filters & EvidenceDrawer
- [x] Sub-tugas 3.3: Virtualized Apps inventory with category filtering, search, and detail drawer
- [x] Sub-tugas 3.4: Interactive Popup Hunter monitor with user marker, telemetry, and candidate actions
- [x] Sub-tugas 3.5: Interactive Audit and Remediation views with safety guardrails, confirmation dialog, and undo
- [x] Sub-tugas 3.6: Interactive Reports and Settings views with allowlist management & root gitignore fix
- [x] Sub-tugas 3.7: Contract Freeze v1.0.0 & gerbang selesai-tahap T3

## Fase 2 (Tahap T2) — riwayat selesai
- [x] Sub-tugas 2.1: Local variable font `@fontsource-variable/inter`, CSS tokens & Zustand theme store
- [x] Sub-tugas 2.2: Konfigurasi `react-i18next` & catalog English (`en.json`) tanpa string hardcode
- [x] Sub-tugas 2.3: Implementasi Core Design System (`Button`, `Card`, `Badge` [icon+label+color], `Skeleton`, `EmptyState`, `ErrorState`, `Table`, `Dialog`, `Tooltip`, `Tabs`)
- [x] Sub-tugas 2.4: Implementasi Aceternity Sidebar (220px/64px, pure CSS 200ms ease-out, zero framer-motion) & registrasi di `docs/ui-components.md`
- [x] Sub-tugas 2.5: Implementasi layout shell, DeviceHeader, StatusBar, dan HashRouter 8 rute
- [x] Sub-tugas 2.6: Implementasi PreflightView dengan alur verifikasi P1–P7, aksi unduh/pilih, dan transisi ke app utama
- [x] Sub-tugas 2.7: Populasi seluruh 8 modul placeholder dengan fixture kontrak dan state handling (loading/empty/error)
- [x] Sub-tugas 2.8: Three.js Scene v0 (`DeviceScene`, `DeviceModel`), fallback SVG (`StaticFallback`), pause on blur/hidden, bounded DPR `[1, 2]`, dan lazy code-split
- [x] Sub-tugas 2.9: Audit checklist PRD §5.6 & gerbang selesai-tahap T2

## Fase 1 (Tahap T1) — riwayat selesai
- [x] Sub-tugas 1.1: Pembaruan aturan proyek & solusi build gitkeep
- [x] Sub-tugas 1.2: Pin Zod & runtime schemas (`schema.ts`)
- [x] Sub-tugas 1.3: Interface bridge & peta event (`bridge.ts`, `events.ts`)
- [x] Sub-tugas 1.4: Fixture JSON bersama (`internal/testdata/contract/*.json`)
- [x] Sub-tugas 1.5: Implementasi mock bridge (`mock.ts`)
- [x] Sub-tugas 1.6: Bridge selector, script `check:no-mock`, & dead-code elimination
- [x] Sub-tugas 1.7: Vitest suite untuk mock lifecycle & events
- [x] Sub-tugas 1.8: Dokumen tata kelola (`docs/contract.md`) & gerbang selesai-tahap T1

## Fase 0 — riwayat selesai
- [x] Git init di branch `main` dengan `.gitattributes` (LF), `.gitignore`, `LICENSE`, `README.md`, `CHANGELOG.md`
- [x] Dokumentasi inti di `docs/` (`docs/PRD.md`, `docs/PROGRESS.md`, ADR 0001–0003)
- [x] Scaffold Wails v2 (React+TS) tanpa folder bersarang
- [x] Struktur folder arsitektur sesuai PRD §6.3
- [x] Backend setup: module path `github.com/EnoughForYou/Jerion`, harness test Go placeholder (sementara, dihapus di T4)
- [x] Frontend setup: Tailwind v3.4.19, tokens CSS, dependencies terkunci (React 19.3.0, Three 0.186.1, R3F 9.8.1, Zustand, Lucide, dll.)
- [x] Tooling: ESLint strict, Prettier (LF), Vitest setup
- [x] Spike WebGL: build Wails + jendela hidup terverifikasi di WebView2 (render visual belum terverifikasi)
- [x] CI workflow GitHub Actions (pin runner Windows untuk wails build, Ubuntu untuk test)
- [x] Acceptance & Gerbang Selesai-Fase 0
- [x] CI hijau di GitHub Actions (Run ID 38147814374: Ubuntu & Windows lolos)

## Keputusan baru (ADR)
- ADR-0001: Stack terkunci (Wails v2.16.0, Go 1.27.1, React 19.3.0, Tailwind v3.4.19, Three 0.186.1, R3F 9.8.1)
- ADR-0002: Kebijakan distribusi ADB dan preflight per-peluncuran (P1–P7)
- ADR-0003: Kebijakan UI English-only dengan lapisan i18n
- ADR-0004: Zod sebagai Single Source of Truth untuk Runtime Schema Kontrak
- ADR-0005: ClamAV Daemon (clamd TCP) Integration for Local Threat Scanning

## Masalah terbuka / blocker
- golangci-lint lokal: Tidak terverifikasi (binary golangci-lint belum terpasang di mesin dev)

## Pengukuran & Kualitas (Tahap T3)
- **Vitest Suite**: 114 pengujian lolos di 19 file tes (`src/features/monitor/monitor.test.tsx`, `src/app/shell.test.tsx`, `src/features/audit/audit.test.tsx`, `src/features/apps/apps.test.tsx`, `src/features/settings/settings.test.tsx`, `src/features/scan/scan.test.tsx`, `src/features/placeholders.test.tsx`, `src/features/reports/reports.test.tsx`, `src/features/preflight/preflight.test.tsx`, `src/components/aceternity/Sidebar.test.tsx`, `src/features/remediate/remediate.test.tsx`, `src/components/ui/components.test.tsx`, `src/three/scene.test.tsx`, `src/App.test.tsx`, `src/lib/bridge/mock.test.ts`, `src/lib/bridge/fixtures.test.ts`, `src/i18n/i18n.test.ts`, `src/app/theme.test.ts`, `src/lib/bridge/schema.test.ts`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker atau skema runtime mock terdeteksi pada bundle produksi Vite.
- **Production Build (`wails build`)**: Kompilasi binary Windows x64 berhasil tanpa error dalam 20.5 detik.
- **Contract Freeze**: Versi kontrak terkunci pada `1.0.0 (FROZEN)` mendefinisikan boundary stabil Go DTO dan Zod TS sebelum Tahap T4.
- **Anti-AI-Slop Checklist (PRD §5.6)**: 100% dipatuhi di seluruh 8 modul interaktif (1 warna aksen, severity berbasis icon+label+warna, font lokal tanpa CDN, zero framer-motion, 1 scene 3D lazy dengan SVG fallback dan DPR `[1, 2]`, teks faktual tanpa lorem).

## Pengukuran & Kualitas (Tahap T2)
- **Vitest Suite**: 92 pengujian lolos di 12 file tes (`src/three/scene.test.tsx`, `src/features/preflight/preflight.test.tsx`, `src/components/aceternity/Sidebar.test.tsx`, `src/components/ui/components.test.tsx`, `src/app/shell.test.tsx`, `src/App.test.tsx`, `src/features/placeholders.test.tsx`, `src/lib/bridge/mock.test.ts`, `src/i18n/i18n.test.ts`, `src/app/theme.test.ts`, `src/lib/bridge/schema.test.ts`, `src/lib/bridge/fixtures.test.ts`).
- **ESLint & TypeScript**: 0 error, 0 warning pada `npm run lint` & `npm run typecheck`.
- **Mock Elimination (`npm run check:no-mock`)**: 0 marker atau skema runtime mock terdeteksi pada bundle produksi Vite.
- **Bundle Splitting**: `DeviceScene` berhasil dipisah secara dinamis (chunk ~923 kB / gzip ~248 kB), bundle aplikasi inti tetap ramping (~481 kB / gzip ~151 kB).
- **Anti-AI-Slop Checklist (PRD §5.6)**: 100% dipatuhi (1 warna aksen, severity berbasis icon+label+warna, font lokal tanpa CDN, zero framer-motion, 1 scene 3D lazy dengan SVG fallback dan DPR `[1, 2]`, teks faktual tanpa lorem).

## Usulan
- —
