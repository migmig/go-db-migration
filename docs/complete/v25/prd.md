# PRD: UI Readability & Wizard Flow Correction (v25)

## 1. Goal
Improve UI text contrast for better readability and ensure the 3-step wizard can be fully reset/restarted after a migration run is completed.

## 2. Problem Statement
- **Low Contrast:** Many labels and descriptions use `text-slate-500` or `text-slate-400`, which is too light against the background, especially in Dark Mode.
- **Restart Loop:** Once a migration reaches Phase 3 (Monitoring), there is no clear path to return to Step 1 (Source/Target config) to start a new session without refreshing the page.

## 3. Requirements
- **Color Audit:** Replace low-contrast grey text (`slate-500/400`) with higher contrast alternatives (`slate-700/600` for Light, `slate-300/200` for Dark) where important information is displayed.
- **Reset Logic:** 
    - The "Close Monitoring" button in `RunStatusPanel` must trigger a full state reset.
    - `App.tsx` must reset the `step` state to `1` when `resetRunState` is called.
- **Visual Feedback:** Ensure active interactive elements are clearly distinguishable from disabled/readonly text.

## 4. Success Criteria
- WCAG-compliant contrast for primary text.
- User can go from Monitoring back to Step 1 with a single click and start a new migration.
- `make offline` (all tests) continues to pass.
