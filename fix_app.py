import re

with open('frontend/src/app/App.tsx', 'r') as f:
    content = f.read()

# 1. Clean up imports
imports_to_remove = [
    'FormEvent, ', 'useRef, ', 'AuthUser, ', 'RuntimeMeta, ', 'DdlEvent, ', 'DiscoverySummary, ',
    'MetricsState, ', 'ReportSummary, ', 'TableRunState, ', 'ValidationState, ', 'WsProgressMsg, ',
    'WsStatus, ', 'createSessionId, '
]
for imp in imports_to_remove:
    content = content.replace(imp, '')

# 2. Remove unused state variables in App.tsx
states_to_remove = [
    r'  const \[tableSearch, setTableSearch\] = useState\(""\);\n',
    r'  const \[tableStatusFilter, setTableStatusFilter\] = useState<TableHistoryStatusFilter>\("all"\);\n',
    r'  const \[tableSort, setTableSort\] = useState<TableSortOption>\("table_asc"\);\n',
    r'  const \[excludeMigratedSuccess, setExcludeMigratedSuccess\] = useState\(false\);\n',
]
for state in states_to_remove:
    content = re.sub(state, '', content)

# 3. Remove setDiscoverySummary(null)
content = content.replace('setDiscoverySummary(null);', '')

# 4. Remove filteredTables, allVisibleSelected, selectedTableSet
filtered_tables_re = r'  const filteredTables = useMemo\(\(\) => \{[\s\S]*?return filtered;\n  \}, \[.*?\]\);\n'
content = re.sub(filtered_tables_re, '', content)

content = re.sub(r'  const selectedTableSet = new Set\(selectedTables\);\n', '', content)
content = re.sub(r'  const allVisibleSelected =[\s\S]*?selectedTableSet\.has\(table\)\);\n', '', content)

# 5. Remove unused functions: toggleTable, selectAllVisibleTables, deselectAllVisibleTables
funcs_to_remove = [
    r'  function toggleTable\(.*?\}',
    r'  function selectAllVisibleTables\(\) \{.*?\}',
    r'  function deselectAllVisibleTables\(\) \{.*?\}'
]
for func_re in funcs_to_remove:
    # Need non-greedy match to closing brace of the function
    content = re.sub(func_re, '', content, flags=re.DOTALL)

# Let's do function removal safer
def remove_function(text, fn_name):
    start_idx = text.find(f"function {fn_name}(")
    if start_idx == -1:
        return text
    line_start = text.rfind('\n', 0, start_idx)
    if line_start != -1:
        start_idx = line_start + 1
    brace_count = 0
    in_block = False
    for i in range(start_idx, len(text)):
        if text[i] == '{':
            brace_count += 1
            in_block = True
        elif text[i] == '}':
            brace_count -= 1
        if in_block and brace_count == 0:
            return text[:start_idx] + text[i+1:]
    return text

for fn in ['toggleTable', 'selectAllVisibleTables', 'deselectAllVisibleTables']:
    content = remove_function(content, fn)

# 6. Update TableSelection props in App.tsx
table_selection_old = r'<TableSelection[\s\S]*?tr=\{tr\}\n\s*/>'
table_selection_new = """<TableSelection
                tr={tr}
                allTables={allTables}
                selectedTables={selectedTables}
                setSelectedTables={setSelectedTables}
                objectGroupModeEnabled={objectGroupModeEnabled}
                previewTables={previewTables}
                discoverySummary={discoverySummary}
                previewObjectGroup={previewObjectGroup}
                previewSequences={previewSequences}
                compareEntries={compareEntries}
                compareFilter={compareFilter}
                setCompareFilter={setCompareFilter}
                compareSearch={compareSearch}
                setCompareSearch={setCompareSearch}
                migrationBusy={migrationBusy}
                selectByCategory={selectByCategory}
                tableProgress={tableProgress}
                historyByTable={historyByTable}
                locale={locale}
                openTableHistory={openTableHistory}
                activeTableHistory={activeTableHistory}
                setActiveTableHistory={setActiveTableHistory}
                tableHistoryBusy={tableHistoryBusy}
                tableHistoryError={tableHistoryError}
                activeHistoryDetail={activeHistoryDetail}
                replayHistory={replayHistory}
              />"""
content = re.sub(table_selection_old, table_selection_new, content)

with open('frontend/src/app/App.tsx', 'w') as f:
    f.write(content)

