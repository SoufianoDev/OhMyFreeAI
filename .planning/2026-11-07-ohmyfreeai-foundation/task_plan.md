# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

Latest verified code head: `beab78a04f4ef9ae1421247983d25da66ffc3986` on `foundation/bootstrap` (2026-11-10).

All four workflow families for this exact commit are complete and successful:

- Root CI [42775712875](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712875): 2/2 jobs passed (Ubuntu and Windows); Linux race detector passed.
- SQLite driver evaluation [42775712758](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712758): 4/4 jobs passed (modernc/ncruces on Ubuntu/Windows).
- Wails platform proof [42775712787](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712787): 2/2 jobs passed (Ubuntu native build/CLI/WebView bridge/database reopen and Windows native build/CLI/executable verification).
- Linux distribution compatibility [42775712919](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712919): 8/8 jobs passed, including Ubuntu 22.04 / legacy GTK3-WebKit41. The previously pending Ubuntu 22.04 job is no longer pending; its actual job completed successfully.

A separate earlier Root CI run on commit `cab00b8e8f1841cd79aed47eb87f708d56127e52` (run [42775485018](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775485018)) did fail on both Ubuntu and Windows in `go test ./...`. The failing regression tests exposed missing production enforcement for Restore journal phase transitions, immutable transaction metadata, candidate identity timing, and refusing to overwrite an invalid existing journal. The follow-up commit `beab78a` implemented these checks in `WriteRestoreJournal`, `RemoveRestoreJournal`, canonicalization and the transition validator; the exact commit then passed Root CI [42775712875](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712875) on both platforms. This was a real code defect and is now fixed and verified. The earlier log-fetch 404 related to the distro run did not indicate a failure; its latest job state is 8/8 success.

The Linux parent-directory durability step is implemented: the no-replace move syncs the already-pinned parent directory FD after rename. Windows source opens are handle-relative and request native `FILE_WRITE_THROUGH`; the source handle is flushed and reused for native handle-bound no-replace rename. **However, a 2026-11-10 primary-documentation re-review found that the public Microsoft contracts cited so far do not explicitly guarantee persistence of the parent-directory namespace entry after `NtSetInformationFile(FileRenameInformationEx)`.** `NtCreateFile` documents write-through for file data writes; File Caching describes metadata caching generally; MS-FSCC defines rename semantics but no durability flag. The previous claim that rename metadata durability was explicitly documented is withdrawn. Windows namespace durability remains an unresolved gate. This blocks not only database artifact moves but also every current `WriteRestoreJournal` intent/completion update, because the Windows journal publisher replaces the fixed journal through the same handle-based rename path. See [RESTORE_RECOVERY_EXECUTOR.md](../../docs/architecture/RESTORE_RECOVERY_EXECUTOR.md). Cross-platform no-follow identity readers and the post-move SHA-256/size helper are present, but the helper is not yet integrated into a transaction executor.

**Immediate prerequisites before the transaction executor:** (1) resolve the Windows namespace-durability contract without an undocumented directory flush or path/overwrite fallback; and (2) resolve the Windows terminal-journal lifecycle. `RemoveRestoreJournal` currently returns `errRestoreJournalRemovalUnsupported` on Windows. The journal has one fixed filename and immutable transaction metadata, so a terminal journal cannot be safely retired for a later transaction. Design a durable multi-transaction journal lifecycle (for example, an append-only or archived journal with crash-safe discovery) that preserves prior evidence; add Linux/Windows failure-injection and process-death tests for journal rollover. Do not bypass this with `os.Remove`, overwrite terminal evidence, or ignore the removal error.

After that gate passes, implement the isolated Restore transaction executor around a fully staged candidate. It must use StoreController's exclusive lease, durable intent-before-mutation journaling, no-replace moves, post-move identity verification, rollback/quarantine retention, verified reopen, and fail-closed behavior. Then add failure-injection and abrupt-process-death/restart tests on Linux and Windows. Do not wire startup recovery or GUI/CLI Restore and do not enable Restore before that matrix passes.

**Delivery warning:** `foundation/bootstrap` is 813 commits ahead of `main` (0 behind) according to the latest comparison. Do not merge the branch wholesale. When the first genuinely testable build is ready, prepare a clean, selective delivery commit/release branch containing only required runtime/build files and essential user documentation; keep planning history, experiments and temporary artifacts out of `main`. No merge is authorized by the current CI result.

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
  - [x] Add Linux/Windows no-follow content identity readers and post-move SHA-256/size verification that preserves mismatched destinations
  - [x] Verify content-identity and post-move verification tests on Ubuntu and Windows CI (Root CI 42772101619)
  - [x] Establish and document the Windows NTFS write-through metadata barrier for the retained-handle no-replace rename (Microsoft NtCreateFile/CreateFile contracts; Windows Root CI 42772101619)
  - [x] Add Linux parent-directory fsync after every durable no-replace namespace mutation (Ubuntu Root CI 42772101619)
  - [x] Extend restoring-previous planning to preserve a candidate still at staging before restoring rollback
  - [ ] Implement the isolated collision-safe filesystem transaction executor (no startup/app wiring)
  - [ ] Integrate journal phases with staged candidate publication, post-move identity checks, verified reopen and rollback
  - [ ] Add Linux/Windows failure-injection and abrupt-process-death/restart coverage
  - [ ] Add the startup recovery coordinator before canonical Store open only after the executor test matrix passes
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
- [ ] Add abrupt-process crash/reopen recovery tests on Linux and Windows
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

Current status snapshot (2026-11-10): the StoreController lifecycle work, Wails/config integration, GUI/CLI backup leases, strict journal v3, identity-aware planner, Linux/Windows no-replace primitives, private state-directory qualification, post-move identity helper, Linux parent-directory fsync and Windows NTFS write-through rename boundary are implemented at the low-level layer. The latest verified code commit `beab78a04f4ef9ae1421247983d25da66ffc3986` passed Root CI [42775712875](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712875) (2/2), SQLite [42775712758](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712758) (4/4), Wails [42775712787](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712787) (2/2), and Linux distributions [42775712919](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712919) (8/8, including Ubuntu 22.04). The transaction executor, automatic rollback/reopen orchestration, process-death recovery, startup coordinator and Restore UI/CLI remain unimplemented; Restore stays disabled. Real-host WebKit sandbox validation, long-duration stress/performance, dependency license inventory, current Colab/model/SSE revalidation, packaging/uninstall and full source review remain open. Remote inference remains policy-gated.

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


## Continuation — Linux no-replace directory durability barrier — 2026-11-09

- [x] Add parent-directory `fsync` after Linux `renameat2(RENAME_NOREPLACE)`, using the already-pinned `openat2` directory descriptor rather than reopening a path.
- [x] Add Linux tests for successful directory-descriptor sync and invalid-descriptor error propagation.
- [ ] Verify the Linux move + fsync and verified-identity tests in a real Ubuntu Root CI job, including race detection.
- [ ] Verify the cross-platform identity/verified-move tests in a real Windows Root CI job.
- [ ] Establish a documented Windows namespace durability barrier with equivalent no-replace and handle-binding guarantees, or keep Windows Restore unavailable.
- The Linux implementation is present but not CI-verified. A sync error occurs after the rename may have changed the namespace; the move returns an error and preserves the destination for journal/recovery inspection.
- Restore remains disabled. Do not start the filesystem executor, startup recovery, or GUI/CLI Restore, and do not merge to main.


## Continuation — CI audit and Linux regression-test restoration — 2026-11-09

- [x] Restore the Linux regression test rejecting a symlinked containing directory while retaining the directory-fsync tests.
- [x] Audit actual completed Root, SQLite-driver, Linux distro, and Wails workflow jobs for commit `6bb79ce73ea558bb07b2d24e889ea0c85f2013fc`.
- [ ] Obtain a new Root CI run where both Ubuntu and Windows matrix jobs are actually materialized and complete.
- [ ] Run the restored symlink-parent regression and Linux directory-sync tests in Ubuntu CI.
- [ ] Establish a documented Windows namespace durability barrier with equivalent no-replace and handle-binding guarantees, or keep Windows Restore unavailable.
- Root CI's only materialized Ubuntu job passed all steps, but the run reports failure and no Windows job was created. SQLite driver evaluation's three materialized jobs passed, including two Windows jobs, but they do not exercise Restore filesystem code. All seven materialized Linux distro jobs passed. Wails workflow failed with zero jobs.
- Restore remains disabled. Do not start the executor, startup recovery, GUI/CLI Restore, or merge to main.


## CI verification — restored Linux regression and cross-platform Restore tests — 2026-11-09

- [x] Run restored Linux symlink-parent regression and Linux directory-sync tests in Ubuntu Root CI; package passed in `go test ./...` and `go test -race ./...`.
- [x] Run Root CI on both Ubuntu and Windows; both jobs completed successfully, including `go test ./...`. Windows SQLite package tests passed.
- [x] Verify all four SQLite driver-evaluation jobs (modernc/ncruces on Ubuntu/Windows); all passed.
- [ ] Complete Wails v3 native build proof on Ubuntu and Windows (run #42682591267 was still in progress at last check).
- [ ] Complete Linux distribution compatibility matrix (run #42682591212 was still in progress at last check).
- [ ] Establish a documented Windows namespace durability barrier with equivalent no-replace and handle-binding guarantees, or keep Windows Restore unavailable.
- Root CI run: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42682590849. SQLite driver evaluation: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42682590841.
- Windows Root tests validate the current Windows identity/verified-move code, not namespace crash durability. Restore remains disabled. Do not start the executor, startup recovery, GUI/CLI Restore, or merge to main.


## Follow-up — Wails Windows native build proof — 2026-11-09

- [x] Wails v3 native application build, Windows CLI dispatch smoke test, and executable verification passed on Windows in run #42682591267.
- [ ] Complete the Ubuntu Wails build and Linux WebView/SQLite smoke tests in run #42682591267.
- [ ] Complete the remaining Linux distribution matrix in run #42682591212.
- Root CI and SQLite driver evaluation remain green for the tested code commit. Windows native build success does not resolve the separate Windows Restore namespace-durability gate.
- Restore remains disabled. No executor, startup recovery, GUI/CLI Restore, or main merge.


## Follow-up — Wails platform proof completed — 2026-11-09

- [x] Complete Wails v3 native build proof on Ubuntu and Windows. Ubuntu live WebView bridge/database reopen smoke test passed; Windows native build, CLI dispatch, and executable verification passed.
- [x] Complete Arch Linux, Fedora 44 GTK4, and Debian 13 desktop compatibility jobs.
- [ ] Complete remaining Linux distribution jobs in run #42682591212 (Ubuntu 22.04, openSUSE Tumbleweed, Fedora legacy GTK3, Ubuntu 24.04, Debian 12; some still active/queued at last check).
- Windows namespace durability remains an independent hard gate. Restore stays disabled; no executor, startup recovery, GUI/CLI Restore, or main merge.


## Verification update — Linux distribution matrix completed — 2026-11-09

- [x] Complete all 8 Linux distribution/GTK-WebKit matrix jobs in run [42682591212](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42682591212). All eight succeeded, including native app build, CLI dispatch, and WebView bridge/database reopen smoke tests.
- [x] Reconfirm branch HEAD after CI: `10b301a32477a53a70f9284e74786c7a6e611ab5`; CI code under test is `f1eb8179a3a95e5cabfa19baebbade03e97b8bf5`, with later commits documentation-only.
- [ ] Establish a documented Windows namespace durability barrier that preserves handle-bound no-replace semantics, or retain Windows Restore as unavailable.
- [ ] Keep Restore disabled until the required Linux and Windows safety gates are both met. Do not start executor, durable journal integration, startup recovery, GUI/CLI Restore, or merge to main before then.

### Windows contract review — 2026-11-09

- `FILE_RENAME_INFORMATION_EX` documents no-replace when the replace flag is omitted, but publishes no write-through/durability flag.
- `MoveFileExW(MOVEFILE_WRITE_THROUGH)` has broad “moved on disk” wording, while its explicit flush guarantee is for copy-and-delete moves. The contract does not unambiguously establish same-volume rename namespace durability; the API is path-based and cannot replace the current handle-bound primitive without weakening source/destination identity guarantees.
- `NtFlushBuffersFileEx` requires write/append access on the handle. `NtCreateFile` prohibits requesting file-write/file-append rights for directories and prohibits generic access rights for directory opens. No documented, ordinary-user directory flush access recipe was found.
- `FLUSH_FLAGS_FILE_DATA_SYNC_ONLY` is invalid for directory handles; `ReplaceFileW` says `REPLACEFILE_WRITE_THROUGH` is unsupported.
- Microsoft Q&A clarification pages were treated as questions, not as authoritative engineering answers. The unresolved point is a missing documented API contract, not a claim about NTFS's internal capabilities.
- No path-based fallback, overwrite-capable fallback, or undocumented directory flush is accepted. Restore remains disabled.


## In-place SQLite restore research — 2026-11-09

- [SQLite Online Backup API](https://www.sqlite.org/c3ref/backup_finish.html) holds a write transaction on the destination for the backup operation. If the operation is finished before `sqlite3_backup_step()` returns `SQLITE_DONE`, SQLite rolls back the incomplete destination transaction. The tests exercise both this rollback path and a completed backup followed by abrupt process termination.
- [SQLite backup documentation](https://www.sqlite.org/backup.html) states that a destination in WAL mode cannot be backed up from a source with a different page size (`SQLITE_READONLY`). The implementation in [SQLite backup.c](https://github.com/sqlite/sqlite/blob/master/src/backup.c) attempts to align destination page size and contains logic to handle different source/destination page sizes when the destination is not WAL.
- [SQLite PERSIST journal mode](https://sqlite.org/atomiccommit.html) retains the rollback journal and commits by overwriting its header with zeros rather than deleting the journal file. This avoids a directory-entry deletion at commit. The [journal_mode pragma documentation](https://www.sqlite.org/pragma.html#pragma_journal_mode) distinguishes PERSIST from WAL: WAL mode persists across reopen; PERSIST must be selected again on the restore connection.
- The test result confirms the target file identity remains unchanged during in-place restore. A 4096-byte target accepts a 1024-byte source in PERSIST mode; after closing/reopening, the target reports 1024-byte pages and the expected data. The same mismatch is rejected in WAL mode without changing the previous target.
- [SQLite Win32 VFS source](https://sqlite.org/src/file/src/os_win.c) shows `winSync` calling `FlushFileBuffers` on the file handle. Unlike the Unix VFS sync path, no parent-directory sync is visible in that Windows routine. PERSIST avoids journal filename deletion during commit, but the first journal-file creation and the durability of its directory entry still need explicit scrutiny. The process-termination tests are not power-loss tests.
- Independent blocker: the current [Windows journal publisher](https://github.com/SoufianoDev/OhMyFreeAI/blob/foundation/bootstrap/internal/infrastructure/sqlite/restore_journal_publish_windows.go) publishes JSON journal updates using `MoveFileExW(MOVEFILE_REPLACE_EXISTING|MOVEFILE_WRITE_THROUGH)`. The Microsoft documentation reviewed in the previous finding does not establish same-volume rename namespace durability. In-place database restore alone does not resolve that journal-publication gap.

### Candidate direction — not approved for implementation

Investigate an in-place restore protocol using SQLite's backup API with a pre-existing PERSIST rollback journal, plus a durable transaction marker that can distinguish the old database from the candidate after restart. Candidate/rollback snapshot creation and transaction-state publication still need a proven no-loss protocol. A database-resident marker or preallocated fixed-slot journal may avoid per-phase path renames, but neither is approved yet.

Do not treat this probe as proof of power-loss safety. Keep Restore disabled and do not start the Restore executor, journal integration, startup recovery, UI/CLI work, or merge until the complete platform gates are met.


## 2026-11-09 23:55Z — Windows relative rename follow-up

- The previous Win32 class-22 rename encoded an absolute destination path and could re-resolve a renamed ancestor after the directory guard check. That approach is removed.
- The Windows move and journal-publish paths now use native `NtSetInformationFile` with `FileRenameInformationEx` class 65, the retained destination directory handle in `RootDirectory`, and a simple relative destination leaf.
- The source is opened from the current path of the pinned directory handle. Its final handle path and the pinned directory handle path are checked again after open; the parent is checked again immediately before rename.
- The source handle keeps `FILE_FLAG_WRITE_THROUGH`; source buffers are flushed before rename. This follows Microsoft's documented NTFS metadata-flush behavior for write-through requests, without claiming stronger hardware/power-loss guarantees.
- Regression tests cover an ancestor rename with a replacement tree, rejection of a decoy source handle, and journal replacement using the pinned directory handle.
- CI on base `0e3e2868e6a74c9fdae698b314c987fe39391ddd` is green: Root CI Ubuntu+Windows; SQLite drivers 4/4; Wails Ubuntu+Windows; Linux distribution matrix 8/8. These results do not cover this new Windows change.
- **Blocking gate:** run and inspect fresh Root CI on the new commit, especially Windows Go tests. Keep Restore disabled and do not begin executor/journal recovery/UI/CLI integration or merge to main until the platform safety gates are met.


## 2026-11-10 00:12Z — Windows Root CI test correction

- Root CI on `79c3e6f3e5b8e6420bb965a112aba8f4e6664be4` passed Ubuntu but Windows failed in two regression-test setup steps: `os.Rename(parent, movedParent)` returned `Access is denied` while a descendant directory handle was open. The failure occurred before the tests exercised rename-to-target behavior.
- Rework the regressions to rename the pinned destination directory itself, then create a replacement at its former path. Existing Windows test coverage already verifies that a retained directory handle can continue resolving children from the original directory after that directory is renamed.
- Verify the rename destination remains attached to the original directory handle and never lands in the replacement tree. Keep a separate decoy-handle rejection test.
- Correct accidental extra indentation in the Windows source-open declarations.
- Base commit's other CI: SQLite 4/4, Wails 2/2, Linux distro matrix 8/8. Fresh Root CI is required for this test-only correction.
- Restore remains disabled; no executor, startup recovery, UI/CLI integration, or merge to main.


## 2026-11-10 00:15Z — Windows directory rename test ordering

- Root CI `42726626677`: Ubuntu succeeded; Windows failed in the new regression when it tried to rename a directory while a file handle inside that directory remained open.
- Microsoft FILE_RENAME_INFORMATION documentation explicitly describes restrictions on renaming directories containing open files. The test must not violate that platform rule.
- Reorder the regression: retain the directory handle, rename the directory before opening its source file, create a replacement directory/decoy at the old path, then open the source from the pinned directory handle and execute the handle-relative rename.
- The decoy validation test uses the same rename-before-open order. No production code changed.
- Run fresh Windows Root CI; do not mark Windows safety verified until Go tests pass.
- Restore remains disabled; no executor, startup recovery, UI/CLI integration, or merge to main.


## 2026-11-10 — Verification and JSC signal handling

- [x] Correct the Windows directory-rename regression ordering to respect Windows' open-file rename restrictions.
- [x] Verify final Windows handle-relative no-replace rename, decoy rejection, and journal relative replacement in Windows Root CI.
- [x] Root CI on `410dbf88d091e012f360fe5f3fe409da05811615`: Ubuntu + Windows SUCCESS; Go vet/tests/build, Python tests, frontend syntax and architecture checks passed; Ubuntu race detection passed.
- [x] SQLite driver evaluation on the same commit: modernc/ncruces × Ubuntu/Windows 4/4 SUCCESS.
- [x] Wails v3.0.0 platform proof on the same commit: Ubuntu + Windows SUCCESS; Linux live bridge and database reopen smoke passed; Windows native build/CLI smoke passed.
- [x] Linux distribution compatibility matrix on ancestor `79c3e6f`: 8/8 SUCCESS; no production code changed in subsequent test/planning commits.
- [x] Investigate `Overriding existing handler for signal 10`: verified Wails v3.0.0's upstream fix leaves JSC's SIGUSR1 handler alone, reapplies Go-compatible SA_ONSTACK flags to other relevant signals during first WebView initialization, and re-applies after page load. No arbitrary `JSC_SIGNAL_FOR_GC` override is warranted.
- [ ] Keep Restore disabled until full namespace durability, post-move identity, journal recovery and platform safety gates are independently complete. CI success is not a power-loss guarantee.
- No Restore executor, startup recovery, UI/CLI Restore, or merge to main.


## 2026-11-10 07:11Z — Restore platform-gate closure and current CI

- Verified branch head `9ca1f7e6de8ecca00970b3201601ca70f31ddb95` on `foundation/bootstrap`; latest commit message: `test(sqlite): use private restore test directories`.
- Root CI [42772101619](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101619): Ubuntu and Windows jobs SUCCESS. Ubuntu passed vet, Go tests, race detector, build, Python compile/tests, frontend syntax and architecture checks. Windows passed vet, Go tests, build, Python compile/tests, frontend syntax and architecture checks; Linux-only race step skipped as expected.
- SQLite driver evaluation [42772101510](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101510): 4/4 modernc/ncruces × Ubuntu/Windows jobs SUCCESS.
- Wails platform proof [42772101659](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101659): 2/2 jobs SUCCESS. Linux native build, CLI smoke, executable verification, WebView bridge and database reopen passed; Windows native build, CLI smoke and executable verification passed.
- Linux distribution compatibility [42772101557](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101557): 8/8 distro jobs SUCCESS (Debian 12/13, Ubuntu 22.04/24.04, Fedora 44 GTK3/GTK4, Arch rolling, openSUSE Tumbleweed).
- Linux no-replace move now fsyncs the pinned parent-directory FD after rename. Windows source open uses native FILE_WRITE_THROUGH, FlushFileBuffers, and the same retained handle for NtSetInformationFile(FileRenameInformationEx=65) with no replacement. Microsoft primary documentation now supports the NTFS metadata-flush request for rename operations. This is not a hardware-independent power-loss guarantee.
- No transaction executor, automatic startup recovery, process-death test matrix, or GUI/CLI Restore path exists. Restore remains disabled.
- GitHub compare reports 804 commits ahead and 0 behind main; no open PR exists. Do not merge the branch wholesale. The first real testable delivery must be curated selectively.
- Next: implement the isolated StoreController-backed transaction executor with intent-before-mutation journaling, no-replace moves, post-move identity verification, rollback/quarantine retention, verified reopen and Linux/Windows process-death tests.
