# PRD: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Goal
Improve the user experience when migrating a large number of tables (100+) and optimize the layout for modern large/wide monitors to utilize available screen real estate effectively.

## 2. Problem Statement
- **Narrow Layout:** On large monitors, the UI is restricted to `max-w-7xl` (1280px), leaving excessive whitespace on the sides and compressing complex components.
- **Table List Scaling:** When selecting or monitoring many tables, the lists either feel too small (Step 2) or grow excessively long (Step 3), making it difficult to maintain a global view.
- **Visual Breakage:** Long table names or many concurrent status updates can cause layout shifting or wrapping that degrades readability.

## 3. Requirements

### 3.1 Fluid Layout for Large Screens
- **Expanded Container:** Increase the maximum container width from `max-w-7xl` to `max-w-[1600px]` (or `max-w-[90vw]`) to give components more room to breathe on wide displays.
- **Responsive Grids:** 
    - In `RunStatusPanel`, increase grid columns for wider screens (e.g., `2xl:grid-cols-4`, `3xl:grid-cols-6`).
    - Adjust `TableSelection` side-by-side layout to use more horizontal space.

### 3.2 Scalable Table Selection (Step 2)
- **Increased Height:** The available/selected table lists should use more vertical space (`h-[500px]` or `min-h-[40vh]`) instead of the current `h-64`.
- **Top-aligned Controls:** Selection buttons (Add/Remove) are moved to the top of each panel for easier access when many tables are present.
- **Multi-term Search:** Support comma-separated search terms in both "Available" and "Selected" table filters (e.g., `TB_USER, TB_ORDER`).
- **Long Name Handling:** Ensure table names truncate gracefully with tooltips or provide enough horizontal width to avoid line breaks.

### 3.3 Optimized Progress Monitoring (Step 3)
- **Scrollable Progress Section:** The "Detailed Table Progress" grid should be contained within a scrollable area (`max-h-[60vh]`) with a sticky header if necessary, preventing the entire page from becoming excessively long.
- **Filtering:** Add a quick filter (search box) specifically for the "Detailed Table Progress" list to quickly locate specific tables among hundreds.
- **Compact View Option:** (Optional/Stretch) Add a toggle to switch between "Card" view and "Row" view for table progress.

### 3.4 Visual Polish
- **Sticky Actions:** Ensure primary actions (Next, Start Migration, Stop) remain easily accessible even when lists are scrolled.
- **Loading States:** Ensure smooth transitions when loading large metadata sets.

## 4. Success Criteria
- UI fills wide screens appropriately without looking "stretched."
- 200+ tables can be selected and monitored without layout breakage or excessive page-level scrolling.
- "Search" and "Filter" functionality remains performant with high table counts.
- `make offline` (all tests) continues to pass.
