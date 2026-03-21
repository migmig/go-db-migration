# Tasks: UI Readability & Wizard Flow (v25)

## Phase 1: Contrast Fixes
- [ ] Audit `App.tsx` and change `slate-500` to `slate-700` (Light) and `slate-400` to `slate-300` (Dark).
- [ ] Audit `TableSelection.tsx` for table names and counts readability.
- [ ] Audit `MigrationOptionsPanel.tsx` for labels and help text.
- [ ] Audit `RunStatusPanel.tsx` for monitoring metrics and table list.

## Phase 2: Flow Fixes
- [ ] Update `App.tsx` `handleResetRunState` to include `setStep(1)`.
- [ ] Ensure `useMigrationRun.ts` `resetRunState` properly clears `runSessionId` and `migrationBusy`.
- [ ] Verify that clicking "Close Monitoring" correctly navigates the user back to Step 1.

## Phase 3: Validation
- [ ] Run `make offline` to ensure no test regressions.
- [ ] Manual check of Step 1 -> 2 -> 3 -> Reset -> Step 1 flow.
