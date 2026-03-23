# Tasks: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Research & Analysis
- [x] Verify Tailwind configuration for custom breakpoints (if any).
- [x] Measure layout performance with 100+ simulated tables.

## 2. Layout Implementation
- [x] Update `App.tsx` container width: `max-w-7xl` -> `max-w-[1600px]`.
- [x] In `TableSelection.tsx`, update table list heights to `h-[500px]` and improve horizontal wrapping.
- [x] In `RunStatusPanel.tsx`, wrap table progress grid in a scrollable container with `max-h-[60vh]`.
- [x] Adjust grid columns for `RunStatusPanel` for `2xl` and `3xl` screens.
- [x] Move Table Selection buttons to the top for better accessibility.

## 3. Performance & Features
- [x] Implement search filter for detailed table progress in `RunStatusPanel.tsx`.
- [x] Support comma-separated multi-term search in Table Selection.
- [x] Add Google Login integration (Backend & Frontend).
- [x] Add Database Connection History saving logic.

## 4. Testing & Validation
- [x] Verify layout on high-resolution displays.
- [x] Test with "large" metadata (100+ tables).
- [x] Ensure mobile view (`sm` and `lg`) is still working as expected.
- [x] Run `make offline` to verify no regressions in unit/integration tests.
