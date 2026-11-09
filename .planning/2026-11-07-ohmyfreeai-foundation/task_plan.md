# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

The baseline workflow 42419106009 passed Ubuntu and Windows on `f5396ee592b9b9756779f67aeabd6d7167520c40`. StoreController integration is already present and verified: Wails composition root, AppService, config repository, GUI/CLI backup leases, shutdown drain, stale-store rejection and failure-safe replacement tests.

Windows Restore path traversal opens each component relative to retained parent handles. The owner/DACL and volume gate is implemented on `foundation/bootstrap`: directory ownership must match either the process TokenUser or its explicit TokenOwner; NULL DACLs and broad writable ACEs are rejected; only fixed local NTFS with persistent ACLs is qualified.

**The gate is now CI-verified.** Root CI [42643886765](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42643886765), commit `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`, passed Ubuntu and Windows. Ubuntu passed vet, Go tests, race detector, build, Python compile/tests, frontend syntax and architecture checks. Windows passed vet, Go tests, build, Python compile/tests, frontend syntax and architecture checks; Linux-only race detection was skipped as expected. Wails [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) also passed both platforms.

The cross-platform no-follow content identity readers and post-move SHA-256/size verification helper are now present in the branch. The helper preserves a moved destination and returns an error on identity mismatch; it does not open SQLite and does not provide a namespace durability barrier. Latest observed branch head: `38ffa21f8b2fef05bc98b80e4ed5c8c1836bd0ff`; no CI status or PR-triggered workflow run was returned for that commit, so these new tests are not yet CI-verified.

**New blocker before executor implementation:** primary Windows API documentation reviewed on 2026-11-09 does not establish a supported non-administrative user-mode durability barrier for the namespace entry changed by the retained-handle `NtSetInformationFile(FileRenameInformationEx=65)` no-replace rename. `FileRenameInformationEx` exposes no write-through flag; `FlushFileBuffers` requires GENERIC_WRITE and only documents volume-wide flushing with administrative privileges; `MoveFileExW(MOVEFILE_WRITE_THROUGH)` does not establish the required guarantee for the same-volume rename path and would abandon the retained-source-handle race protection. Resolve this with primary documentation and a Windows regression test, or keep Restore unavailable on Windows. Linux also still needs parent-directory fsync after each namespace mutation. Do not implement the executor, startup recovery, or GUI/CLI Restore until identity tests and both platform durability contracts are verified.

## Current Phase

Phase 4

## Phases

### Phase 1: Requirements & Discovery

- [x] Recover the agreed v3.3 architecture and project constraints from conversation context
- [x] Inspect the target repository and confirm its starting state
- [x] Inspect required Material Design and Kotlin Compose Kit skill repositories
- [x] Inspect the supplied prototype archives and Python scripts before adopting any behavior
- [x] Record findings and source boundaries
**Status:** complete

### Phase 2: Planning & Structure

- [x] Establish architecture documentation and repository conventions
- [x] Vendor Planning with Files v3.28.1, Material 3 v1.1.1, and Compose component-design guidance under .agents/skills/
- [x] Define module boundaries, ports/adapters, runtime lifecycle, and state contracts
- [x] Define compatibility layers including Colab resource fit and verification levels
- [x] Define omfaibackup format and backward-compatibility policy
**Status:** complete

### Phase 3: Implementation

- [x] Add the framework-independent Go domain foundation and standard-library developer toolkit
- [x] Add validated versioned configuration defaults, repository port and application service
- [x] Add SQLite initial/up and development-only down migrations with schema integrity tests
- [x] Add Linux/Windows GitHub Actions validation workflow
- [x] Add a pinned Wails v3.0.0 official vanilla-js starter build proof to GitHub Actions
- [x] Extend the Wails proof to generate and assert the starter's Go-to-JavaScript GreetService binding on both native runners
- [x] Select Wails v3.0.0 after official and application-specific builds passed on Ubuntu 24.04 and Windows
- [x] Add the actual Wails AppService with GetSnapshot and SaveConfig methods; generate and verify the JavaScript binding
- [x] Add a reproducible Python build script that generates the official Wails template in a temporary workspace and builds the real application
- [x] Select modernc.org/sqlite v1.62.0 after the Linux/Windows driver matrix and production migration SQL passed
- [x] Implement the production SQLite Store, per-connection PRAGMAs, transactional migration runner and validated config repository
- [x] Pin modernc dependency graph in go.mod/go.sum and verify tidy does not change committed lockfiles
- [x] Add an isolated SQLite driver evaluation module and Ubuntu/Windows matrix for modernc and ncruces; keep experimental dependencies out of the production module
- [x] Establish the initial bilingual Material 3 web shell using the repository's Material Design skill
- [x] Preserve historical Colab scripts and launcher behavior unchanged; keep remote-inference activation policy-gated
**Status:** complete

### Phase 4: Testing & Verification

- [x] Run repository Go vet/tests/build, Python tests, frontend syntax and architecture checks on Ubuntu and Windows via GitHub Actions
- [x] Vendor Hush and Deslop Code hygiene skills and make their scoped cleanup rules mandatory in AGENTS.md and Planning with Files
- [x] Evaluate modernc and ncruces on Ubuntu and Windows, including the production initial migration SQL
- [x] Test the production SQLite adapter, migration rollback/idempotence, version-gap and corrupt-history rejection, config round-trip, file permissions and per-connection PRAGMAs on Ubuntu and Windows
- [x] Build the complete Wails application on Ubuntu and Windows; generate 1 service, 2 methods and 7 models, produce the Vite production bundle and verify the native executable
- [x] Pin and verify Go dependency lockfiles in CI
- [x] Upgrade GitHub Actions to Node 24-compatible releases; latest core, SQLite and Wails workflows pass without the Node 20 deprecation warning
- [ ] Run documentation/link validation and review all source changes
- [x] Perform a read-only static audit of the supplied prototype archives and Python patch/monitor scripts; record hashes, preserved invariants and risks in docs/architecture/PROTOTYPE_AUDIT.md
- [ ] Revalidate current Colab /props and /v1/models behavior, model-specific Q8_0/262144 compatibility and the live SSE path before enabling remote inference
- [ ] Reassess the Colab Free policy gate against the official MCP runtime API; distinguish notebook/runtime management from authorization to serve model inference to OpenCode. Keep remote inference disabled until the exact flow is confirmed permitted.
- [x] Add SQLite online backup/restore primitives with integrity, foreign-key, migration-history, cancellation, no-overwrite, target-safety and rollback tests on Ubuntu and Windows
- [x] Implement strict .omfaibackup snapshot archive with manifest, SHA-256, fixed entries, size limits and path-traversal protections; round-trip, rollback, checksum, overwrite and cancellation tests pass on Linux/Windows
- [x] Implement read-only Manager.Inspect with isolated temporary extraction, full manifest/checksum/schema validation and cleanup tests
- [x] Wire backup export/inspect into the Wails service using native file dialogs and bilingual settings UI
- [x] Add CLI backup export/inspect commands using the shared archive manager and existing-only SQLite open
- [x] Specify crash-safe restore lifecycle, durable journal, startup recovery and failure-injection requirements in RESTORE_LIFECYCLE.md
- [ ] Complete coordinated Store lifecycle and Restore safety gates before enabling Restore
  - [x] Complete StoreController TDD and green Ubuntu/Windows root CI
  - [x] Route Wails configuration and backup export through controller leases; remove long-lived raw Store references
  - [x] Route CLI backup export through controller leases
  - [x] Add StoreController shutdown drain and closed-controller rejection tests
  - [x] Publish a verified rollback Store after failed exclusive replacement, or fail closed
  - [x] Add concurrent configuration Load/Save and backup-export drain tests covering complete operations
  - [x] Add concurrent StoreController lease stress test and canceled-shutdown regression test
  - [x] Run the Go race detector in Linux CI; the Windows workflow skips the race step
  - [ ] Add failure-injection coverage across Restore publication, reopen, journal commit and rollback
  - [x] Add strict durable Restore journal v3 storage with bounded decoding, path confinement and paired content identities (low-level only)
  - [x] Record explicit previous-database state and paired candidate/previous content identities in journal v3
  - [x] Add an identity-aware pure recovery planner, intent-phase truth tables and fail-closed ambiguity tests
  - [x] Add Linux/Windows no-replace move primitives; common behavioral tests now run on both platforms
  - [x] Replace Windows path-based MoveFileExW with native NtSetInformationFile(FileRenameInformationEx=65), rooted directory handle and opened source handle
  - [x] Add Windows source-handle substitution and no-replace collision tests; observe RED then GREEN on Windows CI
  - [x] Open Windows source leaf relative to the already-pinned directory handle with NtCreateFile; verify directory-path substitution test on Windows CI
  - [x] Open Windows destination directory component-by-component from the drive root; reject reparse/device/non-directory components and verify pinned-parent behavior on Windows CI
  - [x] Verify canonical state-directory owner and private DACL; fail closed on broad write access (Windows Root CI 42643886765)
  - [x] Qualify the actual volume type and filesystem for the no-replace move; reject unsupported remote/mapped or filesystem types (fixed local NTFS only; Windows Root CI 42643886765)
  - [x] Add Linux/Windows no-follow content identity readers and post-move SHA-256/size verification that preserves mismatched destinations (implementation present; CI verification pending)
  - [ ] Verify the current content-identity and post-move verification tests on Ubuntu and Windows CI
  - [ ] Establish a documented Windows namespace durability barrier for the retained-handle no-replace rename; otherwise keep Windows Restore unavailable
  - [ ] Add Linux parent-directory fsync after every durable no-replace namespace mutation
  - [x] Extend restoring-previous planning to preserve a candidate still at staging before restoring rollback
  - [ ] Implement collision-safe filesystem recovery executor and startup recovery before canonical Store open
  - [ ] Integrate journal phases with staged restore publication, verified reopen and rollback
- [x] Previous native Ubuntu 24.04 WebView-to-Go and two-launch DB reopen smoke passed (Wails run 42404058636)
- [x] Revalidate the refactored two-stage bridge bootstrap on native Ubuntu; run 42405729902 passed with two GetSnapshot calls per launch and DB reopen
- [x] Complete the same WebView-to-Go and two-launch DB reopen smoke on all eight Linux distro jobs (run 42405827037)
- [x] Diagnose GitHub runner WebKit bubblewrap namespace restriction and keep WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS=1 confined to the smoke-test subprocess only
- [ ] Validate WebKit's production sandbox on a real Linux host; container CI explicitly does not certify sandbox behavior
- [x] Define explicit Linux distribution targets and add GTK4/WebKitGTK 6.0 plus legacy GTK3/WebKit2GTK 4.1 build/runtime smoke matrix (Ubuntu 24.04, Debian 13, Fedora 44, Arch, openSUSE Tumbleweed, Ubuntu 22.04, Debian 12, Fedora 44 legacy)
- [x] Verify all eight Linux distro jobs pass in run 42405827037; distinguish container userland coverage from bare-metal certification
- [x] Implement framework-independent GUI-default CLI parser with typed unknown-command/extra-argument errors and unit tests
- [x] Verify the built native executable handles `--help`, `--version`, `doctor`, unknown commands and read-only doctor behavior on Ubuntu 24.04 and Windows (Wails run 42409974970)
- [x] Verify native CLI smoke on all eight Linux distro lanes (run 42410214638)
- [ ] Add abrupt-process crash/reopen recovery tests
- [x] Diagnose Windows 8.3/long-path mismatch in journal round-trip and normalize canonical paths
- [x] Replace fixed Xvfb display :97 with dynamically allocated display numbers in the Linux WebView smoke helper
- [x] Record exact pass/fail results; do not equate compilation with runtime certification
**Status:** in_progress

### Phase 5: Delivery

- [ ] Review branch diff and ensure no unrelated/destructive changes
- [ ] Summarize implemented foundation, verification, and remaining blockers
- [ ] Provide the branch and next concrete engineering step
**Status:** pending

## Key Questions

1. What files and implementation already exist in the repository beyond the README?
2. Which exact Material Design 3 skill assets and Kotlin Compose Kit guidance are available?
3. Which parts of the historical Colab prototype are reproducible and safe to preserve?
4. Which Wails/Go/Colab combinations can be verified without introducing paid dependencies?

Resolved so far: Wails v3.0.0 native builds and live WebView bridge/SQLite reopen passed on Ubuntu 24.04 and Windows in run 42415564460. That same run passed the full native backup CLI end-to-end smoke on both platforms: expected no-database/missing-archive errors, successful export and inspect, exact two-entry archive, overwrite rejection and source DB preservation. Root CI run 42415960280 passed Ubuntu and Windows on commit c4d7bf5c, including Go tests, Python tests and the 60-second WebView smoke helper syntax/tests. The current Wails retest for the 60-second timeout change is run 42415960222 (in progress); the current eight-distro matrix is run 42415960198 (queued). The previous backup-enabled distro matrix 42412554358 passed 7/8; Ubuntu 22.04 GTK3 failed twice at the old 30-second bridge timeout, while the same lane passed in 42405827037. Ubuntu 24.04 Wails logs showed approximately 26 seconds per WebView launch, so the smoke timeout was raised to 60 seconds and Xvfb readiness is now socket-probed with Xvfb logs captured. Restore remains disabled; a dedicated [RESTORE_LIFECYCLE.md](RESTORE_LIFECYCLE.md) now specifies the exclusive Store lease, durable journal, staging verification, rollback and startup crash-recovery requirements. The StoreController now provides shared/exclusive lifecycle leases and is wired into Wails configuration plus GUI/CLI backup export. Durable journal/recovery and application-level rollback/reopen still do not satisfy the full contract. The CI-only WebKit sandbox bypass remains confined to smoke-test subprocesses; real-host production sandbox validation remains open. Still open: implement and failure-test the Store lifecycle/journal before restore, abrupt-crash recovery, long-duration stress/performance, dependency license inventory, current Colab/model/SSE revalidation, packaging/uninstall and full source review. Colab remote-inference activation remains policy-gated.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Work on foundation/bootstrap | Isolates the initial foundation from the default branch and avoids rewriting main prematurely |
| Free-only remains a hard constraint | Explicit project requirement; no paid services or alternative hosting suggestions |
| Preserve the Colab + llama.cpp + Gateway + OpenCode path until audited | It is the only proven direction in the existing project history |
| Backup extension/name is omfaibackup | Explicit user decision |
| Use Planning with Files structure under .planning | Persistent task state, isolated from future tasks |
| Use native Ubuntu/Windows CI plus distro-container compatibility lanes | Ubuntu 24.04, Debian 13, Fedora 44, Arch rolling and openSUSE Tumbleweed exercise GTK4/WebKitGTK 6.0; Ubuntu 22.04, Debian 12 and Fedora 44 legacy exercise the Wails v3.0.x GTK3/WebKit2GTK 4.1 path. Container results are not bare-metal certification |
| Select Wails v3.0.0 as the desktop shell | The actual OhMyFreeAI app, AppService binding and Vite production bundle build on Ubuntu 24.04 and Windows; live WebView invocation and packaging remain unverified |
| Select modernc.org/sqlite v1.62.0 for the SQLite adapter | Both CGO-free candidates passed Linux/Windows smoke tests and the initial migration; modernc avoids the WASM VFS and documented per-connection memory overhead of ncruces |
| Do not claim the skill is installed in the agent runtime | The available environment has no writable agent-home mount; repository vendoring is the durable project-level route |
| Keep Wails confined to the desktop adapter | Core domain/application packages remain Wails-independent; the desktop build is generated from the pinned official vanilla-js template in a temporary workspace |
| Keep the historical Colab monitor read-only | Preserve the busy-session guard and SSE framing invariant, but do not port Linux-only Wails v2 UI/process code into the cross-platform product |
| Mandatory Ponytail Ultra engineering mode | Challenge unjustified scope before implementation; prefer the smallest complete change while preserving security, validation, data-loss protections, user constraints and required tests |

## Errors Encountered

| Error | Attempt | Resolution |
|-------|---------|------------|
| Git clone failed: DNS resolution for github.com unavailable in the container | 1 | Use the connected GitHub integration for repository reads/writes; do not repeat the network clone |
| Code Mode tool-call limit exceeded during bulk skill vendoring | 1 | Reconcile partial writes and split into smaller batches |
| GitHub create-file returned HTTP 422 because a path already existed | 1 | Fetch the existing file SHA and use update_file |
| Policy-gate Go test failed after adding required policy revision | 1 | Add the missing revision to the test fixture and rerun the full Go suite |
| Browser blocked file:// and localhost navigation | 1 | Use Playwright page.set_content() with inline local assets |
| Playwright locator matched both nav and quick-access buttons | 1 | Target unique data-page selectors and rerun successfully |
| Planning phase status markers initially included a list bullet | 1 | Corrected to the upstream **Status:** value format and manually revalidated all five phases |
| Initial Wails build script passed the project folder as the CLI destination parent | 1 | Read the Wails init output, pass a workspace parent to -d, and rebuild the real app successfully on both OSes |
| Go test source contained a literal newline inside a quoted format string | 1 | Escaped the newline and reran Go vet/tests/build on Ubuntu and Windows |
| SQL splitter source template emitted malformed quote/newline rune literals | 1 | Replaced ambiguous escaped byte literals with numeric byte constants and reran the full Go matrix |
| Wails smoke run 42399602513 failed its Ubuntu WebView readiness assertion after 12 seconds; output only showed a DRI3/EGL warning | 1 | Do not assume the warning is causal. Replace stdout grep with an explicit opt-in marker written by GetSnapshot, force software rendering under Xvfb, and wait for the actual JS-to-Go invocation |
| Windows Wails build printed a non-fatal "-VThe system cannot find the file specified." line | 1 | Native executable and binding generation still succeeded; source of the message is not yet identified and remains a tracked warning |

## Notes

- Repository source of truth: https://github.com/SoufianoDev/OhMyFreeAI
- Planning skill source: https://github.com/OthmanAdi/planning-with-files
- The exact Planning with Files initializer could not be executed in this container because GitHub DNS/network access is unavailable and connector files are not mounted into the shell. Preserve the shipped template structure and use the permitted manual fallback; validate the format with the available validator source when possible.


## Continuation — platform no-replace restore move primitive

- Added a test-first contract for moveRestoreFileNoReplace: regular-file move, destination collision preserving both files, symlink/non-regular source rejection, same-directory restriction, and concurrent destination collision (exactly one winner).
- Confirmed RED on test-only commit 61a19ea5932bbda9ef722e8789c5af42cabe4c45: Root Go CI run 42438137306, Ubuntu job 128601331157, failed at go vet with the expected undefined symbol.
- Implemented shared validation plus Linux renameat2(RENAME_NOREPLACE), Windows MoveFileEx(MOVEFILE_WRITE_THROUGH) without replacement/copy flags, and an explicit unsupported-platform fail-closed implementation.
- Declared existing golang.org/x/sys v0.49.0 as a direct dependency; no new module version was introduced.
- Latest code commit: d717eac7096d98c3b59b1b9ebd39aa1cff6e2d93.
- Root Go CI 42438368020: Ubuntu + Windows passed; Linux race detector passed. SQLite driver matrix 42438368022: 4/4 passed. Wails 42438368038 and distro matrix 42438368003 are pending; Ubuntu 22.04 from the previous matrix is being retried separately.
- Next: close the pending Wails/distro checks, then implement the durable executor around these primitives with intent-before-mutation journaling, identity revalidation, directory sync, process-death tests and idempotent recovery. Do not wire startup recovery or enable Restore until all safety gates pass.


### CI update after no-replace primitive

- Root Go CI 42438368020 passed Ubuntu + Windows; Linux race detector passed.
- SQLite driver evaluation 42438368022 passed 4/4.
- Wails native platform proof 42438368038 passed 2/2, including Ubuntu WebView-to-Go and SQLite reopen.
- Previous distro matrix 42437541325 reached 8/8 after rerunning only its failed Ubuntu 22.04 GTK3 lane; that targeted retry passed without a code change.
- Current code-revision distro matrix 42438368003 remains pending and must be checked separately.


## Continuation — platform-specific journal namespace barriers

- Removed the false Windows success path from the shared directory-sync helper.
- Non-Windows journal publication now owns its complete sequence: rename then parent-directory sync.
- Windows journal publication now uses same-directory MoveFileExW with REPLACE_EXISTING + WRITE_THROUGH after syncing the temporary file.
- Windows journal deletion fails closed and preserves the journal; no production caller currently depends on removal.
- Tests were added first in commits f876c42 and db2eddd, but their workflow attempts were cancelled before test execution. No RED evidence is claimed.
- Implementation commits: a51e958, 45c5329, cbccd640, a68bba1, 1cfd080, 673782c, 48bc940.
- Latest code CI is pending: Ubuntu root job queued, Windows root job cancelled; Wails and distro jobs queued. Do not claim pass.
- Next: close current CI, then design and implement the executor only after the platform durability boundary is explicit. Windows remains unqualified for full power-loss Restore until NTFS-specific process-death/restart evidence exists.


### Verification closure — journal namespace barriers

- Code revision 6b582b3f1ee5537fb4f5c15c209e8fe02a75bc68 passed Root Go CI run 42439943545: Ubuntu SUCCESS and Windows SUCCESS.
- Ubuntu passed go vet, go test, go test -race, build, Python compile/tests, frontend syntax and architecture checks.
- Windows passed go vet, go test, build, Python compile/tests, frontend syntax and architecture checks; Linux-only race step was skipped as expected.
- Eight-distro matrix run 42439943571 remained queued when recorded; no distro-matrix success is claimed for this revision.
- Test-first commits for journal namespace changes were cancelled before test execution. No RED evidence is claimed for that subtask.


## Continuation — Windows handle-bound no-replace rename — 2026-11-08

- Reviewed Microsoft's MS-FSCC FileRenameInformationEx specification and the Win32 SetFileInformationByHandle contract, plus the exact pinned x/sys v0.49.0 API surface.
- Added a Windows-only test for the critical identity property: after opening the candidate handle, rename its old path away and create a replacement at that path; the operation must publish the already-opened original, not the replacement.
- Added a Windows-only destination-collision test that verifies both source and destination bytes remain unchanged.
- Observed RED on test-only commit b398c546a3737f689d0ce7be9204f20b97e4b627: Root CI run 42496137044, Windows job 128778788124, failed at go vet with the expected undefined renameRestoreHandleNoReplace symbol.
- Implemented Windows FileRenameInfoEx class 22 through SetFileInformationByHandle. The source is opened with DELETE and FILE_FLAG_OPEN_REPARSE_POINT; source and destination directory handles are retained; reparse points and non-regular sources are rejected; RootDirectory is the destination directory handle; replacement flag is zero; the variable UTF-16 buffer uses the documented 20-byte fixed header and 24-byte minimum structure size.
- Implementation commit: 18aa3a3455857d5a68cd80afeb92a674ca6d1eca. Root CI 42496232166, Wails 42496232229, SQLite driver evaluation 42496232217 and distro matrix 42496232194 are pending at the time of this note; no GREEN result is claimed yet.
- The handle-bound rename helper is a low-level primitive only. The path-based source open still depends on the future executor's private-state-directory and path-confinement guarantees; Windows parent-component reparse traversal is not yet independently pinned through a relative NT open.
- No volume capability probe or filesystem qualification gate is wired yet. GitHub Windows runner success, if achieved, is a test result on that runner's volume, not general NTFS/ReFS/SMB certification.
- Restore remains disabled. No executor, journal orchestration, directory durability, process-death recovery, startup recovery, or app-level Restore service has been added.


## Correction — Windows rooted rename requires the native NT API — 2026-11-08

- The first handle-bound implementation used SetFileInformationByHandle(FileRenameInfoEx=22) with a non-null RootDirectory. Windows Root CI run 42496232166 passed Ubuntu but failed Windows: the regular move, collision and concurrent-collision tests all returned ERROR_INVALID_PARAMETER ("The parameter is incorrect").
- This was not a buffer-size or compilation failure. Microsoft's Win32 SetFileInformationByHandle documentation's supported-class table lists FileRenameInfo (3), not FileRenameInfoEx. A source-handle rename using a rooted relative destination is supported at the native NtSetInformationFile layer.
- Exa research verified Microsoft's FILE_INFORMATION_CLASS table: FileRenameInformationEx is native class 65, available starting with Windows 10 version 1709. Microsoft's NtSetInformationFile documentation describes the native operation and its information-class buffer. A practitioner source-code case also documents that passing non-null RootDirectory through SetFileInformationByHandle is rejected with ERROR_INVALID_PARAMETER.
- Corrected the adapter to call x/sys/windows.NtSetInformationFile with native class 65, the existing source handle, retained destination directory handle, and the MS-FSCC 24-byte-minimum FILE_RENAME_INFORMATION_EX payload. No replacement flag is set. STATUS_OBJECT_NAME_COLLISION maps to os.ErrExist.
- Corrected implementation commit: d6f6bb54c9f72cfcd118087e4468e7d425e83251.
- Current validation runs for this corrected commit: Root CI 42496496153, Wails 42496496135, SQLite driver evaluation 42496496199, and Linux distro matrix 42496496136. They were pending at the time of this note; no GREEN claim.
- No legacy FileRenameInformation fallback is added. NTFS is the initial target; filesystems that reject class 65 remain unsupported. The source path is still opened by pathname and parent-component reparse traversal is not pinned; the future executor must enforce a private canonical state directory. Restore remains disabled.


## CI closure — Windows handle-bound move — 2026-11-08

- Root CI run 42496659593 passed Ubuntu and Windows. Ubuntu passed go vet, go test, go test -race, build, Python compile/tests, frontend syntax and architecture checks. Windows passed go vet, go test, build, Python compile/tests, frontend syntax and architecture checks; race is Linux-only.
- SQLite driver evaluation run 42496659645 passed 4/4: modernc and ncruces on Ubuntu 24.04 and Windows.
- Wails platform proof run 42496659732 passed 2/2. Ubuntu passed native build, CLI smoke, executable verification, two-launch WebView-to-Go bridge and database reopen; Windows passed native build and executable verification.
- Linux distribution matrix run 42496659764: Arch Linux, Fedora 44 GTK4, Ubuntu 24.04 GTK4 and Debian 13 GTK4 passed. openSUSE Tumbleweed, Ubuntu 22.04 GTK3, Fedora 44 legacy GTK3 and Debian 12 GTK3 were still running at this update. Do not claim 8/8 yet.
- Current code/test commit: c85c3350d6dd9c0557805be62a6761f1a8110129. The Windows handle-bound implementation itself is in d6f6bb54c9f72cfcd118087e4468e7d425e83251.
- Restore remains disabled. The move primitive still opens the destination directory by path and the source by pathname; next close the source-open race by using NtCreateFile with RootDirectory set to the already-opened directory handle. The future executor must also enforce a private canonical state directory and journal/recovery invariants.


## Final CI closure — Windows handle-bound move — 2026-11-08

- Linux distribution matrix 42496659764 completed SUCCESS, 8/8: Arch rolling, Fedora 44 GTK4, Ubuntu 24.04 GTK4, Debian 13 GTK4, openSUSE Tumbleweed, Ubuntu 22.04 GTK3, Fedora 44 legacy GTK3 and Debian 12 GTK3.
- The full code/test revision c85c3350 is green across Root CI, Wails, SQLite driver evaluation and the eight-lane Linux matrix:
  - Root CI 42496659593: Ubuntu + Windows SUCCESS; Linux race detector SUCCESS.
  - Wails 42496659732: Ubuntu + Windows SUCCESS; Ubuntu WebView-to-Go and DB reopen passed.
  - SQLite drivers 42496659645: 4/4 SUCCESS.
  - Linux distributions 42496659764: 8/8 SUCCESS.
- The current branch also contains the separately committed Wails v3.1 compatibility ADR index update and this Restore CI record. No production code changed after c85c3350.


## Windows relative source open — 2026-11-08

- Added a Windows-only test that opens the state-directory handle, renames that directory to a different path, creates a replacement directory at the old path, then requires the source open and rename to operate inside the original pinned directory.
- Observed RED on test-only commit 3f1110029423cdf5893732f086af8d385dd02b8d: Root CI run 42497376480 / Windows job 128782981057 failed at go vet with the expected undefined openRestoreSourceRelative symbol.
- Implemented source opening through x/sys/windows.NtCreateFile with OBJECT_ATTRIBUTES.RootDirectory set to the retained destination directory handle. It opens one validated leaf component with FILE_OPEN_REPARSE_POINT | FILE_NON_DIRECTORY_FILE | FILE_SYNCHRONOUS_IO_NONALERT and requests SYNCHRONIZE, DELETE and FILE_READ_ATTRIBUTES. This prevents re-resolving the source parent path after the destination directory handle has been opened.
- The first implementation attempt passed nil for NtCreateFile's uintptr EA-buffer argument and failed go vet. Corrected to zero in commit c89d532a3ba04d43c7d2ccbed1675319c6f90ee8.
- Root CI run 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed. SQLite driver evaluation 42497620420 passed 4/4.
- Wails run 42497620429 was still in progress and Linux distribution matrix 42497620422 was queued at this update.
- Remaining boundary: the destination directory itself is still opened by a full path with CreateFile. OPEN_REPARSE_POINT protects its final component, not every ancestor. Next, open the canonical state path component-by-component from a trusted volume root, reject reparse points at each step, and verify private directory permissions/ACL. The executor must still hash and revalidate identity from the opened source handle before mutation.
- Restore remains disabled.


## Current code revision verification — c89d532a — 2026-11-08

- Root CI 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed. The new Windows relative-source-open regression test ran as part of the Windows Go package suite.
- SQLite driver evaluation 42497620420 passed 4/4.
- Wails platform proof 42497620429 passed 2/2, including Linux WebView-to-Go/database reopen and Windows native executable verification.
- Linux distribution matrix 42497620422 was in progress with 4/8 jobs passed at this update. Do not claim full matrix green for c89d532a yet.
- Code commit: c89d532a3ba04d43c7d2ccbed1675319c6f90ee8. Current branch also includes docs-only follow-up commits.
- Remaining Windows path boundary: destination directory path is opened by CreateFile; OPEN_REPARSE_POINT protects its final component only. Next implement component-by-component directory opening from a trusted volume root and verify private ACL/permissions before the executor.


## 2026-11-09 CI gate update

- Root CI `42642834055` failed only on Windows Go package tests; Ubuntu completed successfully. Windows failed because the restore-directory owner check compared the filesystem owner exclusively to TokenUser.
- Wails `42642834036` passed on Ubuntu and Windows, including native build/CLI checks and Ubuntu WebView bridge/database reopen.
- Corrected the Windows identity check to accept only TokenUser or the current token's explicit TokenOwner (the documented default owner for newly created objects). The broad-writable DACL rejection and fixed-local-NTFS gate remain unchanged.
- Fix commit: `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`; planning update commit: `c1d9063b90283a40a5668b85f252f9bf37ac64b9`.
- **Next mandatory gate:** rerun Root CI on this revision and inspect both platform results. Do not treat the fix as verified until Windows tests pass. Keep Restore disabled; do not start executor/journal/recovery/UI before the platform safety gate is green.


### CI closure — Windows Restore directory security gate — 2026-11-09

- Root CI [42643886765](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42643886765) passed on commit `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`: Ubuntu and Windows SUCCESS.
- Ubuntu passed Go vet/tests, race detector, build, Python compile/tests, frontend syntax and architecture boundaries. Windows passed Go vet/tests/build, Python compile/tests, frontend syntax and architecture boundaries; the Linux-only race step was skipped as designed.
- This verifies the Windows TokenUser/TokenOwner owner comparison correction, private DACL checks, broad-principal write rejection, and fixed-local-NTFS/persistent-ACL volume qualification on the Windows CI runner. It does not qualify other Windows filesystem providers or power-loss behavior.
- Wails [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) passed Ubuntu and Windows native build/smoke checks.
- Code fix: `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`. No changes were made to `main`; Restore remains disabled.


## Latest CI status — verified-move tests — 2026-11-09 18:15Z

- [ ] Root CI run [42671657858](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671657858) did not provide a test result: run conclusion is FAILURE, but the jobs endpoint returned zero jobs and the Actions page said “Waiting for pending jobs”. No code-level failure is established.
- [ ] Identity helper tests remain unverified on both platforms. The SQLite driver workflow's only created job (Windows modernc) passed, but it does not run the main Restore package; the overall run reports FAILURE with expected matrix cells absent.
- [ ] Wails Windows native build/CLI smoke passed, but the overall workflow reports FAILURE with the Ubuntu lane absent. Linux distribution compatibility remained in progress on Debian 12 native build.
- [ ] Re-run/obtain actual Ubuntu and Windows Root CI Go test results for the identity reader and verified-move tests. Do not infer success from the Windows-only ancillary workflows.
- [ ] Resolve the workflow-level pending-job/scheduling state before another code push; avoid superseding the active Linux distribution run with a no-op commit.
- [ ] Linux parent-directory fsync and the documented Windows namespace durability barrier remain blocking prerequisites for executor work.
- Restore remains disabled; no executor, startup recovery, app-level Restore service, CLI/GUI Restore entry point, or merge to main.


## CI closure — verified-move revision — 2026-11-09 18:16Z

- [ ] Root CI [42671657858](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671657858) remains unverified: run failed before any jobs were returned; no Go tests ran.
- [ ] Verified-move identity tests remain unverified on Linux and Windows.
- [x] Partial platform evidence: Windows Wails native build/CLI smoke passed in run [42671658474](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658474), but the Ubuntu matrix lane was not created and the workflow summary is FAILURE.
- [x] Partial Linux evidence: Debian 12 legacy GTK3/WebKit2GTK 4.1 native build, CLI smoke, WebView bridge, and database reopen passed in run [42671658275](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658275). Only 1/8 matrix lanes ran; overall workflow is FAILURE.
- [x] Partial SQLite evidence: Windows modernc migration/driver tests passed in run [42671658166]; the other expected matrix cells were absent and the overall workflow is FAILURE. This is not evidence for the main Restore package.
- [ ] Resolve the missing GitHub Actions matrix jobs and obtain a real Ubuntu + Windows Root CI run before marking the new tests verified.
- [ ] Implement and test Linux parent-directory fsync; establish a documented Windows namespace durability barrier or keep Windows Restore unavailable.
- Restore remains disabled. No executor, startup recovery, app-level Restore service, CLI/GUI Restore, or main merge.
