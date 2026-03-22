# Tasks: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Research & Analysis
- [ ] Verify Tailwind configuration for custom breakpoints (if any).
- [ ] Measure layout performance with 100+ simulated tables.

## 2. Layout Implementation
- [ ] Update `App.tsx` container width: `max-w-7xl` -> `max-w-[1600px]`.
- [ ] In `TableSelection.tsx`, update table list heights to `h-[500px]` and improve horizontal wrapping.
- [ ] In `RunStatusPanel.tsx`, wrap table progress grid in a scrollable container with `max-h-[60vh]`.
- [ ] Adjust grid columns for `RunStatusPanel` for `2xl` and `3xl` screens.

## 3. Performance & Features
- [ ] Implement search filter for detailed table progress in `RunStatusPanel.tsx`.
- [ ] Optimize rendering of table progress cards (React memo if needed).

## 4. Testing & Validation
- [ ] Verify layout on high-resolution displays.
- [ ] Test with "large" metadata (100+ tables).
- [ ] Ensure mobile view (`sm` and `lg`) is still working as expected.
- [ ] Run `make offline` to verify no regressions in unit/integration tests.
