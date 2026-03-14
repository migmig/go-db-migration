import re

with open("internal/web/templates/index.html", "r", encoding="utf-8") as f:
    content = f.read()

# Replace Step 1 Connect content
step1_old = """    <!-- ── Step 1: Connect ── -->
    <div class="step-content active" id="step-1">
        <div id="conn-section" class="card">
            <h2>Oracle DB Connection</h2>
            <div class="form-group">
                <label for="oracleUrl">Oracle URL</label>
                <input type="text" id="oracleUrl" placeholder="localhost:1521/XE" required aria-label="Oracle URL">
            </div>
            <div style="display: flex; gap: 1rem; margin-bottom: 1.25rem;">
                <div class="form-group" style="flex: 1; margin-bottom: 0;">
                    <label for="username">Username</label>
                    <input type="text" id="username" required aria-label="Oracle 사용자명">
                </div>
                <div class="form-group" style="flex: 1; margin-bottom: 0;">
                    <label for="password">Password</label>
                    <input type="password" id="password" required aria-label="Oracle 패스워드">
                </div>
            </div>
            <div class="form-group">
                <label for="likeQuery">Table Filter (LIKE)</label>
                <input type="text" id="likeQuery" placeholder="e.g. USERS_%" aria-label="테이블 필터 (LIKE 패턴)">
            </div>
            <button id="btn-connect" class="btn-primary" aria-label="Oracle 연결 및 테이블 조회">Connect &amp; Fetch Tables</button>
            <div id="conn-error" class="error-msg" role="alert"></div>
            <div id="conn-success" class="conn-success-banner" style="display:none" role="status"></div>
        </div>
    </div>"""

step1_new = """    <!-- ── Step 1: Connect ── -->
    <div class="step-content active" id="step-1">
        <div class="step2-grid">
            <!-- Source DB Panel -->
            <div id="conn-section" class="card" style="margin-bottom: 0;">
                <h2>1. Source DB Connection (Oracle)</h2>
                <div class="form-group">
                    <label for="oracleUrl">Oracle URL</label>
                    <input type="text" id="oracleUrl" placeholder="localhost:1521/XE" required aria-label="Oracle URL">
                </div>
                <div style="display: flex; gap: 1rem; margin-bottom: 1.25rem;">
                    <div class="form-group" style="flex: 1; margin-bottom: 0;">
                        <label for="username">Username</label>
                        <input type="text" id="username" required aria-label="Oracle 사용자명">
                    </div>
                    <div class="form-group" style="flex: 1; margin-bottom: 0;">
                        <label for="password">Password</label>
                        <input type="password" id="password" required aria-label="Oracle 패스워드">
                    </div>
                </div>
                <div class="form-group">
                    <label for="likeQuery">Table Filter (LIKE)</label>
                    <input type="text" id="likeQuery" placeholder="e.g. USERS_%" aria-label="테이블 필터 (LIKE 패턴)">
                </div>
                <button id="btn-connect" class="btn-primary" aria-label="Oracle 연결 및 테이블 조회">Connect &amp; Fetch Tables</button>
                <div id="conn-error" class="error-msg" role="alert"></div>
                <div id="conn-success" class="conn-success-banner" style="display:none" role="status"></div>
            </div>

            <!-- Target DB Panel -->
            <div id="target-conn-section" class="card" style="margin-bottom: 0;">
                <h2>2. Target DB Connection</h2>
                
                <h3 class="settings-section-title" style="margin-bottom:0.75rem;">출력 방식 (Migration Mode)</h3>
                <div class="form-group">
                    <label style="display: flex; align-items: center; gap: 0.5rem; font-weight: 400; cursor: pointer;">
                        <input type="radio" name="migrationMode" id="modeFile" value="file" checked aria-label="SQL 파일 생성 모드">
                        SQL 파일 생성
                    </label>
                    <label style="display: flex; align-items: center; gap: 0.5rem; font-weight: 400; cursor: pointer; margin-top: 0.5rem;">
                        <input type="radio" name="migrationMode" id="modeDirect" value="direct" aria-label="Direct Migration 모드">
                        Direct Migration (SQL 파일 없이 직접 이관)
                    </label>
                </div>

                <div id="target-db-controls" style="display: none; margin-top: 1.5rem; border-top: 1px solid var(--border-color); padding-top: 1.5rem;">
                    <div class="form-group">
                        <label for="targetDb">대상 DB</label>
                        <select id="targetDb" onchange="handleTargetDbChange()" aria-label="대상 데이터베이스 선택">
                            <option value="postgres" selected>PostgreSQL</option>
                            <option value="mysql">MySQL</option>
                            <option value="mariadb">MariaDB</option>
                            <option value="sqlite">SQLite</option>
                            <option value="mssql">MSSQL (SQL Server)</option>
                        </select>
                    </div>
                    <div id="pg-config" style="animation: fadeIn 0.3s ease-in-out;">
                        <div class="form-group">
                            <label for="pgUrl" id="targetUrlLabel">대상 URL</label>
                            <input type="text" id="pgUrl" placeholder="postgres://user:pass@host:5432/dbname" aria-label="대상 DB 연결 URL">
                        </div>
                    </div>
                    <div class="form-group" id="schema-group">
                        <label for="schema">스키마</label>
                        <input type="text" id="schema" placeholder="public" aria-label="대상 스키마">
                    </div>
                    <button id="btn-test-target" class="btn-secondary" aria-label="타겟 DB 연결 테스트">Test Target Connection</button>
                    <div id="target-error" class="error-msg" role="alert"></div>
                    <div id="target-success" class="conn-success-banner" style="display:none" role="status"></div>
                </div>
            </div>
        </div>
        <div style="display: flex; justify-content: flex-end; margin-top: 1.5rem; margin-bottom: 2rem;">
            <button id="btn-next-step1" class="btn-primary" style="width: auto; min-width: 120px;" disabled aria-label="다음 단계">Next &rarr;</button>
        </div>
    </div>"""

content = content.replace(step1_old, step1_new)

# Replace Step 2 right panel up to DDL Options
step2_old = """                <!-- Section A: Output mode -->
                <h3 class="settings-section-title">A. 출력 방식</h3>
                <div class="form-group">
                    <label style="display: flex; align-items: center; gap: 0.5rem; font-weight: 400; cursor: pointer;">
                        <input type="radio" name="migrationMode" id="modeFile" value="file" checked aria-label="SQL 파일 생성 모드">
                        SQL 파일 생성
                    </label>
                    <label style="display: flex; align-items: center; gap: 0.5rem; font-weight: 400; cursor: pointer; margin-top: 0.5rem;">
                        <input type="radio" name="migrationMode" id="modeDirect" value="direct" aria-label="Direct Migration 모드">
                        Direct Migration (SQL 파일 없이 직접 이관)
                    </label>
                </div>
                <div id="sql-only-controls" style="margin-top: 0.5rem;">
                    <div class="form-group" style="margin-bottom: 0.75rem;">
                        <label for="outFile">출력 파일명</label>
                        <input type="text" id="outFile" value="migration.sql" aria-label="출력 파일명">
                    </div>
                    <div style="display: flex; gap: 2rem;">
                        <div class="checkbox-container" style="margin-bottom: 0;">
                            <input type="checkbox" id="perTable" checked aria-label="테이블별 개별 파일 출력">
                            <label for="perTable">테이블별 개별 파일 출력</label>
                        </div>
                        <div class="checkbox-container" style="margin-bottom: 0;">
                            <input type="checkbox" id="logJson" aria-label="JSON 로깅 활성화">
                            <label for="logJson">JSON 로깅</label>
                        </div>
                    </div>
                </div>

                <hr class="settings-divider">

                <!-- Section B: Target DB -->
                <h3 class="settings-section-title">B. 대상 데이터베이스</h3>
                <div class="form-group">
                    <label for="targetDb">대상 DB</label>
                    <select id="targetDb" onchange="handleTargetDbChange()" aria-label="대상 데이터베이스 선택">
                        <option value="postgres" selected>PostgreSQL</option>
                        <option value="mysql">MySQL</option>
                        <option value="mariadb">MariaDB</option>
                        <option value="sqlite">SQLite</option>
                        <option value="mssql">MSSQL (SQL Server)</option>
                    </select>
                </div>
                <div id="pg-config" style="display: none; animation: fadeIn 0.3s ease-in-out;">
                    <div class="form-group">
                        <label for="pgUrl" id="targetUrlLabel">대상 URL</label>
                        <input type="text" id="pgUrl" placeholder="postgres://user:pass@host:port/dbname" aria-label="대상 DB 연결 URL">
                    </div>
                </div>
                <div class="form-group" id="schema-group">
                    <label for="schema">스키마</label>
                    <input type="text" id="schema" placeholder="public" aria-label="대상 스키마">
                </div>

                <hr class="settings-divider">

                <!-- Section C: DDL options -->
                <h3 class="settings-section-title">C. DDL 옵션</h3>"""

step2_new = """                <!-- Section A: Output file -->
                <h3 class="settings-section-title" id="title-a">A. 출력 파일 설정</h3>
                <div id="sql-only-controls">
                    <div class="form-group" style="margin-bottom: 0.75rem;">
                        <label for="outFile">출력 파일명</label>
                        <input type="text" id="outFile" value="migration.sql" aria-label="출력 파일명">
                    </div>
                    <div style="display: flex; gap: 2rem;">
                        <div class="checkbox-container" style="margin-bottom: 0;">
                            <input type="checkbox" id="perTable" checked aria-label="테이블별 개별 파일 출력">
                            <label for="perTable">테이블별 개별 파일 출력</label>
                        </div>
                        <div class="checkbox-container" style="margin-bottom: 0;">
                            <input type="checkbox" id="logJson" aria-label="JSON 로깅 활성화">
                            <label for="logJson">JSON 로깅</label>
                        </div>
                    </div>
                </div>

                <hr class="settings-divider" id="divider-a">

                <!-- Section B: DDL options -->
                <h3 class="settings-section-title">B. DDL 옵션</h3>"""

content = content.replace(step2_old, step2_new)

# Replace "D. 고급 설정" to "C. 고급 설정"
content = content.replace("D. 고급 설정", "C. 고급 설정")

# Javascript modifications

js_dom_old = """    const btnConnect      = document.getElementById('btn-connect');
    const btnMigrate      = document.getElementById('btn-migrate');
    const pgConfig        = document.getElementById('pg-config');"""

js_dom_new = """    const btnConnect      = document.getElementById('btn-connect');
    const btnTestTarget   = document.getElementById('btn-test-target');
    const btnNextStep1    = document.getElementById('btn-next-step1');
    const btnMigrate      = document.getElementById('btn-migrate');
    const pgConfig        = document.getElementById('pg-config');"""

content = content.replace(js_dom_old, js_dom_new)

js_state_old = """    let currentZipId  = '';
    let isDryRunMode  = false;"""

js_state_new = """    let currentZipId  = '';
    let isDryRunMode  = false;
    let sourceConnected = false;
    let targetConnected = false;
    let isDirectMode    = false;"""

content = content.replace(js_state_old, js_state_new)

js_reset_old = """        document.getElementById('migrate-error').innerText = '';
        connSuccess.style.display = 'none';
        btnMigrate.disabled = false;"""

js_reset_new = """        document.getElementById('migrate-error').innerText = '';
        connSuccess.style.display = 'none';
        sourceConnected = false;
        targetConnected = false;
        document.getElementById('target-error').innerText = '';
        document.getElementById('target-success').style.display = 'none';
        updateNextButtonState();
        btnMigrate.disabled = false;"""

content = content.replace(js_reset_old, js_reset_new)

js_radio_old = """    // ── Migration mode radio toggle ──
    document.querySelectorAll('input[name="migrationMode"]').forEach(radio => {
        radio.addEventListener('change', (e) => {
            const isDirect = e.target.value === 'direct';
            pgConfig.style.display        = isDirect ? 'block' : 'none';
            sqlOnlyControls.style.display = isDirect ? 'none'  : 'block';
        });
    });"""

js_radio_new = """    // ── State management ──
    function updateNextButtonState() {
        if (sourceConnected && (!isDirectMode || targetConnected)) {
            btnNextStep1.disabled = false;
        } else {
            btnNextStep1.disabled = true;
        }
    }

    btnNextStep1.addEventListener('click', () => {
        showStep(2);
        document.querySelector('.container').scrollTo({ top: 0, behavior: 'smooth' });
    });

    // ── Migration mode radio toggle ──
    document.querySelectorAll('input[name="migrationMode"]').forEach(radio => {
        radio.addEventListener('change', (e) => {
            isDirectMode = e.target.value === 'direct';
            const targetDbControls = document.getElementById('target-db-controls');
            if (targetDbControls) targetDbControls.style.display = isDirectMode ? 'block' : 'none';
            
            const dividerA = document.getElementById('divider-a');
            const titleA = document.getElementById('title-a');
            if (sqlOnlyControls) sqlOnlyControls.style.display = isDirectMode ? 'none'  : 'block';
            if (dividerA) dividerA.style.display = isDirectMode ? 'none' : 'block';
            if (titleA) titleA.style.display = isDirectMode ? 'none' : 'block';

            updateNextButtonState();
        });
    });

    document.getElementById('pgUrl').addEventListener('input', () => {
        targetConnected = false;
        document.getElementById('target-success').style.display = 'none';
        updateNextButtonState();
    });

    document.getElementById('targetDb').addEventListener('change', () => {
        targetConnected = false;
        document.getElementById('target-success').style.display = 'none';
        updateNextButtonState();
    });

    // ── Test Target Connection ──
    btnTestTarget.addEventListener('click', async () => {
        const targetDb  = document.getElementById('targetDb').value;
        const targetUrl = document.getElementById('pgUrl').value.trim();
        const errorDiv  = document.getElementById('target-error');
        const successDiv= document.getElementById('target-success');
        
        errorDiv.innerText = '';
        successDiv.style.display = 'none';

        if (!targetUrl) {
            errorDiv.innerText = 'Please provide a Target DB connection URL.';
            return;
        }

        btnTestTarget.disabled = true;
        btnTestTarget.innerHTML = '<span style="opacity:0.8;">Testing...</span>';

        try {
            const res = await fetch('/api/test-target', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ targetDb, targetUrl })
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.error || 'Failed to connect to target DB');

            targetConnected = true;
            successDiv.innerHTML = `<svg style="width:16px;height:16px;flex-shrink:0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg> Target DB Connected`;
            successDiv.style.display = 'flex';
            updateNextButtonState();
        } catch (err) {
            targetConnected = false;
            errorDiv.innerText = err.message;
            updateNextButtonState();
        } finally {
            btnTestTarget.disabled = false;
            btnTestTarget.innerText = 'Test Target Connection';
        }
    });"""

content = content.replace(js_radio_old, js_radio_new)

js_connect_old = """            // Show success banner
            connSuccess.innerHTML = `<svg style="width:16px;height:16px;flex-shrink:0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg> Connected: <strong>${oracleUrl}</strong> &mdash; ${tables.length} tables found`;
            connSuccess.style.display = 'flex';

            // Advance to step 2 after brief pause
            setTimeout(() => {
                showStep(2);
                document.querySelector('.container').scrollTo({ top: 0, behavior: 'smooth' });
            }, 600);

        } catch (err) {
            errorDiv.innerText = classifyConnError(err.message);
        }"""

js_connect_new = """            // Show success banner
            connSuccess.innerHTML = `<svg style="width:16px;height:16px;flex-shrink:0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg> Connected: <strong>${oracleUrl}</strong> &mdash; ${tables.length} tables found`;
            connSuccess.style.display = 'flex';
            sourceConnected = true;
            updateNextButtonState();

        } catch (err) {
            sourceConnected = false;
            updateNextButtonState();
            errorDiv.innerText = classifyConnError(err.message);
        }"""

content = content.replace(js_connect_old, js_connect_new)


with open("internal/web/templates/index.html", "w", encoding="utf-8") as f:
    f.write(content)

