# SPEC: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Introduction
This spec details the technical changes required to support 4K/Ultrawide monitors and large metadata sets (100+ tables) in the migration UI.

## 2. Structural Layout Adjustments

### 2.1 Main Container (`App.tsx`)
- Change: `max-w-7xl` (1280px) to `max-w-[1600px]` in Step 2 & 3.
- Reasoning: 1280px is too narrow for large displays, making side-by-side table lists feel cramped. 1600px provides a better balance between line length and space utilization.

### 2.2 Table Selection Component (`TableSelection.tsx`)
- Change: Height of available/selected table containers from `h-64` to `h-[500px]`.
- Change: Relocate "Add All/Selected" and "Remove All/Selected" buttons to the top of their respective panels (Left/Right) for better visibility.
- Change: Update search filtering to split input by `,` and perform `OR` matching on multiple terms.
- Change: Added `truncate` and `hover:whitespace-normal` (or tooltips) for table names.
- Change: Ensure horizontal scroll in table body to prevent long names from pushing buttons off-screen.

### 2.3 Run Status Panel (`RunStatusPanel.tsx`)
- Change: Added `max-h-[60vh] overflow-auto` container around the table progress grid.
- Change: Responsive grid adjustment:
    ```tsx
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 3xl:grid-cols-6">
    ```
- Feature: Added a search input above the progress grid to filter tables by name.

## 3. UI/UX Refinement
- Step indicators should remain centered.
- Summary cards in Phase 3 should use more horizontal space to avoid excessive stacking.

## 4. Technical Configuration

### 4.1 Tailwind Configuration (`tailwind.config.ts`)
Add `3xl` breakpoint:
```ts
theme: {
  extend: {
    screens: {
      '3xl': '1920px',
    },
    // ...
  }
}
```

## 5. Non-Functional Requirements
- **Performance:** Filtering 500+ tables in-memory must be instantaneous (< 16ms).
- **Accessibility:** Ensure all new filters and scrollable areas are keyboard-navigable.
