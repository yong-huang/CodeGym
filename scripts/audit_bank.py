#!/usr/bin/env python3
"""CodeGym 题库全面审计：结构 / span / 提示 / 可执行题输出 / 重复项"""
import json, glob, os, re, subprocess, sys, tempfile, shutil
from collections import Counter, defaultdict

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'backend'))
BANK = os.path.join(ROOT, 'data/bank')
VALID_TYPES = {'fill', 'command', 'config', 'bugfix'}
WEIGHTS = {'beginner', 'easy', 'medium', 'hard', 'expert'}
FILL_TYPES = {'text', 'select', 'code', 'code_block'}

E, W = defaultdict(list), defaultdict(list)   # ERROR / WARN

def err(tag, msg):  E[tag].append(msg)
def warn(tag, msg): W[tag].append(msg)

def norm(s): return re.sub(r'\s+', '', s or '')

def occurrence_index(s, sub, n):
    """mimic backend occurrenceIndex"""
    idx = -1
    for _ in range(n + 1):
        nxt = s.find(sub, idx + 1)
        if nxt < 0:
            return None
        idx = nxt
    return idx

# ---------- load ----------
exercises = []
for f in sorted(glob.glob(os.path.join(BANK, '*.json'))):
    name = os.path.basename(f)
    try:
        data = json.load(open(f, encoding='utf-8'))
    except Exception as ex:
        err('json', f'{name}: 解析失败 {ex}')
        continue
    for ex in data:
        ex['_file'] = name
        exercises.append(ex)

print(f'共加载 {len(exercises)} 题（{len(glob.glob(BANK+"/*.json"))} 个文件）')

# ---------- global dup ids ----------
ids = Counter(e.get('id') for e in exercises)
for i, c in ids.items():
    if c > 1 and not i.startswith('fa-'):  # fa- 前缀曾跨文件? no—检查所有
        pass
dupes = {i: c for i, c in ids.items() if c > 1}
if dupes:
    for i, c in dupes.items(): err('dup-id', f'id={i} 出现 {c} 次')

def run_python(code, timeout=15):
    try:
        r = subprocess.run(['python3', '-c', code], capture_output=True, text=True, timeout=timeout, cwd=tempfile.gettempdir())
        return r.stdout.strip(), r.stderr.strip(), r.returncode
    except subprocess.TimeoutExpired:
        return '', 'TIMEOUT', -1

go_dir = None
def run_go(src):
    """compile+run a standalone go program; returns (stdout, ok)"""
    global go_dir
    if go_dir is None:
        go_dir = tempfile.mkdtemp(prefix='bankaudit-go-')
        open(os.path.join(go_dir, 'go.mod'), 'w').write('module audit\n\ngo 1.22\n')
    d = tempfile.mkdtemp(prefix='run-')
    open(os.path.join(d, 'main.go'), 'w').write(src)
    try:
        r = subprocess.run(['go', 'run', os.path.join(d, 'main.go')], capture_output=True, text=True,
                           timeout=60, cwd=go_dir)
        # `go run file.go` works without module per-file; simpler: run directly
        return r.stdout.strip(), r.stderr.strip(), r.returncode
    except subprocess.TimeoutExpired:
        return '', 'TIMEOUT', -1
    finally:
        shutil.rmtree(d, ignore_errors=True)

py_n = go_n = 0
for ex in exercises:
    tag = f"[{ex['_file']}:{ex.get('id','?')}]"
    et = ex.get('exercise_type')
    ts = ex.get('tech_stack')
    w = ex.get('weight')

    # ---- structure ----
    for k in ('id', 'title', 'description', 'exercise_type', 'tech_stack', 'weight', 'category', 'full_code'):
        if not ex.get(k): err('field', f'{tag} 缺字段 {k}')
    if et not in VALID_TYPES: err('enum', f'{tag} exercise_type={et}')
    if w not in WEIGHTS: warn('enum', f'{tag} weight={w}')
    if ts == 'fastapi': err('stack', f'{tag} tech_stack 仍为 fastapi')

    pts = ex.get('points') or []
    if not pts: err('points', f'{tag} 无训练点')
    pids = Counter(p.get('id') for p in pts)
    for pid, c in pids.items():
        if c > 1: err('points', f'{tag} 训练点 id 重复: {pid}')

    code = ex.get('buggy_code') if et == 'bugfix' else ex.get('full_code', '')
    spans = []
    seen_span_text = {}
    for p in pts:
        pt_tag = f"{tag}:p[{p.get('id')}]"
        t = p.get('text')
        if not t: err('points', f'{pt_tag} text 为空')
        mw = p.get('min_weight')
        if mw not in WEIGHTS: err('enum', f'{pt_tag} min_weight={mw}')
        ft = p.get('fill_type', 'text')
        if ft not in FILL_TYPES: err('enum', f'{pt_tag} fill_type={ft}')
        ws = p.get('weight_score', 0)
        if not ws or ws <= 0: warn('score', f'{pt_tag} weight_score={ws}')
        occ = p.get('occurrence', 0) or 0
        start = occurrence_index(code or '', t or '', occ)
        if start is None:
            err('span', f'{pt_tag} text={t!r}(occ={occ}) 不在代码中')
            continue
        end = start + len(t or '')
        spans.append((start, end, p.get('id')))
        key = (start, end)
        if key in seen_span_text:
            err('overlap', f'{tag} 两点同 span: {seen_span_text[key]} 与 {p.get("id")} text={t!r}')
        else:
            seen_span_text[key] = p.get('id')
    # pairwise overlap (partial allowed? backend SelectPoints skips overlaps at selection; still flag full-containment duplicates only via same-span above; partial overlap flag as WARN)
    spans.sort()
    for a, b in zip(spans, spans[1:]):
        if a[1] > b[0]:
            warn('overlap', f'{tag} 部分 overlap: {a[2]}({a[0]}-{a[1]}) vs {b[2]}({b[0]}-{b[1]})')

    # ---- hints ----
    h = ex.get('hints') or {}
    raw = json.dumps(h, ensure_ascii=False)
    for lv in ('level_1', 'level_2', 'level_3', 'level_4'):
        if not h.get(lv): warn('hint', f'{tag} 缺 {lv}')
    if '正确答案：正确答案' in raw: err('hint', f'{tag} L4 重复前缀')
    l4 = h.get('level_4') or ''
    if l4 and not l4.startswith('📝 正确答案：'):
        warn('hint', f'{tag} L4 非标准前缀: {l4[:30]}')
    if not ex.get('explanation'): warn('expl', f'{tag} 缺解析')

    # ---- executable checks ----
    runnable_py = ts == 'python' and et in ('fill', 'bugfix') and h.get('_am') != 'x'
    is_fastapi_am = ts == 'python' and (ex.get('subcategory') == 'FastAPI')
    if ts == 'python' and et in ('fill', 'bugfix') and not is_fastapi_am:
        py_n += 1
        exp = ex.get('expected') or ''
        if not exp:
            err('exec', f'{tag} python 可执行题 expected 为空')
        else:
            src = ex['full_code']
            if et == 'bugfix':
                src += ex.get('validation_code') or ''
                # buggy must NOT produce expected
                b_out, b_err, b_rc = run_python(ex.get('buggy_code') or '')
                buggy_bad = (norm(b_out) != norm(exp)) or b_rc != 0
                if not buggy_bad:
                    err('bugfix', f'{tag} buggy_code 竟然输出与预期一致（不是 bug）')
                tc = ex.get('test_cases') or []
                if not tc: warn('bugfix', f'{tag} 无测试用例')
            out, e_, rc = run_python(src)
            if rc != 0:
                err('exec', f'{tag} 运行失败: {e_.splitlines()[0][:90] if e_ else rc}')
            elif norm(out) != norm(exp):
                err('exec', f'{tag} 输出不符: got={out!r} want={exp!r}')
    if ts == 'go' and ex['_file'] in ('legacy.json', 'go.json'):
        go_n += 1
        exp = ex.get('expected') or ''
        out, e_, rc = run_go(ex['full_code'])
        if rc != 0:
            err('exec-go', f'{tag} 编译/运行失败: {(e_ or "").splitlines()[0][:90]}')
        elif exp and norm(out) != norm(exp):
            err('exec-go', f'{tag} 输出不符: got={out!r} want={exp!r}')

print(f'运行验证: python {py_n} 题, legacy go {go_n} 题')

# ---------- DB-only exercises ----------
import sqlite3
con = sqlite3.connect(os.path.join(ROOT, 'data/codegym.db'))
rows = con.execute('SELECT id,title,exercise_type,tech_stack,subcategory,full_code,buggy_code,validation_code,expected_output,test_cases,fill_points FROM exercises').fetchall()
bank_ids = set(ids)
db_only = [r for r in rows if r[0] not in bank_ids]
print(f'DB 手工题: {len(db_only)} 道')
for rid, title, et, ts, sub, fc, bc, val, exp, tcj, fpj in db_only:
    tag = f'[DB:{rid}]'
    if et in ('fill', 'bugfix') and ts == 'python' and sub != 'FastAPI':
        src = fc + (val or '')
        out, e_, rc = run_python(src)
        if rc != 0:
            err('db', f'{tag} 运行失败: {(e_ or "").splitlines()[0][:80]}')
        elif exp and norm(out) != norm(exp):
            err('db', f'{tag} 输出不符 got={out!r} want={exp!r}')

# ---------- summary ----------
print('\n========== 审计结果 ==========')
for sev in ('ERROR', 'WARN'):
    for cat in E if sev == 'ERROR' else W:
        items = (E if sev == 'ERROR' else W)[cat]
        print(f'\n## {sev} [{cat}] x{len(items)}')
        for m in items:
            print('  -', m)
total_e = sum(len(v) for v in E.values())
total_w = sum(len(v) for v in W.values())
print(f'\nERROR={total_e}  WARN={total_w}')
