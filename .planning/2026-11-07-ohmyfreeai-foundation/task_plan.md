# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

Observe the Wails v3.0.0 native CI proof on Ubuntu 24.04 and Windows; fix any runner/template issues. The workflow now generates and checks the vanilla-js template's GreetService JavaScript binding before building. Keep Wails undecided until both OS jobs pass. In parallel, run the existing Go/Python/architecture checks and evaluate SQLite drivers with an explicit cross-platform test matrix.

## Current Phase

Phase 3

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
- [x] Add a pinned Wails v3.0.0 official vanilla-js starter build proof to GitHub Actions; do not lock the app to Wails until native jobs pass
- [x] Extend the Wails proof to generate and assert the starter's Go-to-JavaScript GreetService binding on both native runners
- [x] Establish the initial bilingual Material 3 web shell using the repository's Material Design skill
- [x] Preserve historical Colab scripts and launcher behavior unchanged; keep remote-inference activation policy-gated
**Status:** in_progress

### Phase 4: Testing & Verification

- [ ] Run repository, documentation, Go, and Python checks available in this environment
- [ ] Verify the original prototype path and identify unverified assumptions
- [ ] Record exact pass/fail results; do not equate compilation with runtime certification
**Status:** pending

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

Resolved so far: Go core tests passed in the previously tested local mirror; the local container cannot run the current Go 1.27.3 module or Wails v3.0.0. A native Ubuntu 24.04/Windows Wails CI proof has now been added, but it is not verified until GitHub Actions reports both jobs green. Colab remote-inference activation remains policy-gated.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Work on foundation/bootstrap | Isolates the initial foundation from the default branch and avoids rewriting main prematurely |
| Free-only remains a hard constraint | Explicit project requirement; no paid services or alternative hosting suggestions |
| Preserve the Colab + llama.cpp + Gateway + OpenCode path until audited | It is the only proven direction in the existing project history |
| Backup extension/name is omfaibackup | Explicit user decision |
| Use Planning with Files structure under .planning | Persistent task state, isolated from future tasks |
| Do not claim the skill is installed in the agent runtime | The available environment has no writable agent-home mount; repository vendoring is the durable project-level route |
| Keep Wails version undecided | v3 is current stable but needs Go 1.26+/GTK4/WebKitGTK 6; v2 has a different platform baseline. Verify on Linux and Windows before selecting |

## Errors Encountered

| Error | Attempt | Resolution |
|-------|---------|------------|
| Git clone failed: DNS resolution for github.com unavailable in the container | 1 | Use the connected GitHub integration for repository reads/writes; do not repeat the network clone |
| Code Mode tool-call limit exceeded during bulk skill vendoring | 1 | Reconcile partial writes and split into smaller batches |
| GitHub create-file returned HTTP 422 because a path already existed | 1 | Fetch the existing file SHA and use update_file |
| Policy-gate Go test failed after adding required policy revision | 1 | Add the missing revision to the test fixture and rerun the full Go suite |
| Browser blocked file:// and localhost navigation | 1 | Use Playwright page.set_content() with inline local assets |
| Playwright locator matched both nav and quick-access buttons | 1 | Target unique data-page selectors and rerun successfully |
| Planning phase status markers initially included a list bullet | 1 | Corrected to the upstream `**Status:** value` format and manually revalidated all five phases |

## Notes

- Repository source of truth: https://github.com/SoufianoDev/OhMyFreeAI
- Planning skill source: https://github.com/OthmanAdi/planning-with-files
- The exact Planning with Files initializer could not be executed in this container because GitHub DNS/network access is unavailable and connector files are not mounted into the shell. Preserve the shipped template structure and use the permitted manual fallback; validate the format with the available validator source when possible.
