import re

with open('frontend/src/app/App.tsx', 'r') as f:
    content = f.read()

# 1. Fix imports - keep useRef, remove unused
content = content.replace(
    'import { FormEvent, useEffect, useMemo, useRef, useState } from "react";',
    'import { useEffect, useMemo, useRef, useState } from "react";'
)

# 2. Fix circular dependency using a ref for resetRunState
content = content.replace(
    '  const [currentStep, setCurrentStep] = useState<1 | 2 | 3>(1);',
    '  const [currentStep, setCurrentStep] = useState<1 | 2 | 3>(1);\n  const resetRunStateRef = useRef<() => void>(() => {});'
)

use_auth_call = """
  const {
    meta,
    user,
    booting,
    bootError,
    loginForm,
    loginBusy,
    loginError,
    setLoginForm,
    boot,
    handleLogin,
    handleLogout,
  } = useAuth({
    resetRunState: () => resetRunStateRef.current(),
    setCredentialsPanelOpen,
    setHistoryPanelOpen,
    setNotice,
  });

  const {
    tableProgress,
    validation,
    ddlEvents,
    warnings,
    discoverySummary,
    metrics,
    migrationBusy,
    migrationError,
    wsStatus,
    runSessionId,
    runStartedAt,
    runEndedAt,
    runDryRun,
    zipFileId,
    reportSummary,
    clock,
    startMigration,
    resetRunState,
  } = useMigrationRun({
    options,
    source,
    target,
    selectedTables,
    effectiveObjectGroup: isObjectGroupModeEnabled(meta) ? options.objectGroup : "all",
    setNotice,
  });
  resetRunStateRef.current = resetRunState;
"""

# Replace the existing hook calls
content = re.sub(r'  const \{\s*meta,[\s\S]*?\}\s*=\s*useAuth\(\{[\s\S]*?\}\);', '', content)
content = re.sub(r'  const \{\s*tableProgress,[\s\S]*?resetRunState,?\s*\}\s*=\s*useMigrationRun\(\{[\s\S]*?\}\);', use_auth_call, content)

# 3. Remove unused state variables
states_to_remove = [
    r'  const \[tableSearch, setTableSearch\] = useState\(""\);\n',
    r'  const \[tableStatusFilter, setTableStatusFilter\] = useState<TableHistoryStatusFilter>\("all"\);\n',
    r'  const \[tableSort, setTableSort\] = useState<TableSortOption>\("table_asc"\);\n',
    r'  const \[excludeMigratedSuccess, setExcludeMigratedSuccess\] = useState\(false\);\n',
]
for s in states_to_remove:
    content = re.sub(s, '', content)

# 4. Remove setDiscoverySummary(null)
content = content.replace('setDiscoverySummary(null);', '')

# 5. Remove filteredTables, allVisibleSelected, selectedTableSet, toggleTable, etc.
def remove_block(text, start_pattern, end_pattern):
    return re.sub(start_pattern + r'[\s\S]*?' + end_pattern, '', text)

content = remove_block(content, r'  const filteredTables = useMemo\(', r'  \}, \[allTables, excludeMigratedSuccess, historyByTable, tableSearch, tableSort, tableStatusFilter\]\);')
content = remove_block(content, r'  const selectedTableSet = new Set\(selectedTables\);', r'    filteredTables\.every\(\(table\) => selectedTableSet\.has\(table\)\);')

def remove_function(text, fn_name):
    start_idx = text.find(f"function {fn_name}(")
    if start_idx == -1: return text
    line_start = text.rfind('\n', 0, start_idx)
    if line_start != -1: start_idx = line_start + 1
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

# 6. Fix boot() call in JSX
content = content.replace('onClick={() => void boot()}', 'onClick={() => void boot()}') # already correct if boot is in scope

with open('frontend/src/app/App.tsx', 'w') as f:
    f.write(content)
