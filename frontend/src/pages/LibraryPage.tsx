import { useEffect, useMemo, useRef, useState } from 'react';
import { api } from '../api/client';
import { TECH_EMOJI, TYPE_LABEL } from '../tech';
import type { CategoryItem, ExerciseDetail, ExerciseListItem } from '../types';

const EMPTY_EXERCISE: ExerciseDetail = {
  id: '',
  title: '',
  description: '',
  exercise_type: 'fill',
  tech_stack: 'go',
  weight: 'medium',
  category: '',
  subcategory: '基础',
  full_code: '',
  fill_points: { points: [] },
  spotter_hints: { level_1: '', level_2: '', level_3: '', level_4: '' },
  time_limit_seconds: 30,
  enabled: true,
};

export default function LibraryPage() {
  const [exercises, setExercises] = useState<ExerciseListItem[]>([]);
  const [categories, setCategories] = useState<CategoryItem[]>([]);
  const [tech, setTech] = useState<string>('');
  const [filterSub, setFilterSub] = useState('');
  const [search, setSearch] = useState('');
  const [newSub, setNewSub] = useState('');
  const [msg, setMsg] = useState('');
  const [err, setErr] = useState('');
  // AI 生成
  const [aiTech, setAiTech] = useState('python');
  const [aiWeight, setAiWeight] = useState('easy');
  const [aiKnowledge, setAiKnowledge] = useState('');
  const [aiCount, setAiCount] = useState(3);
  const [aiBusy, setAiBusy] = useState(false);
  const [aiResults, setAiResults] = useState<{ id: string; title: string; ok: boolean; error?: string }[]>([]);
  // LLM 设置
  const [llmEndpoint, setLlmEndpoint] = useState('');
  const [llmModel, setLlmModel] = useState('');
  const [llmModels, setLlmModels] = useState<string[]>([]);
  const [llmOnline, setLlmOnline] = useState(false);
  // edit/create modal
  const [editing, setEditing] = useState<ExerciseDetail | null>(null);
  const [json, setJson] = useState('');
  const fileRef = useRef<HTMLInputElement>(null);

  const refresh = async () => {
    const [exs, cats] = await Promise.all([api.listExercises(), api.categories()]);
    setExercises(exs);
    setCategories(cats);
    if (!tech && exs.length) {
      // default to the tech with the most questions
      const byTech = new Map<string, number>();
      for (const e of exs) byTech.set(e.tech_stack, (byTech.get(e.tech_stack) ?? 0) + 1);
      const top = [...byTech.entries()].sort((a, b) => b[1] - a[1])[0][0];
      setTech(top);
    }
  };

  useEffect(() => {
    refresh().catch((e) => setErr(String(e)));
    api
      .llmSettings()
      .then((s) => {
        setLlmEndpoint(s.endpoint);
        setLlmModel(s.model);
        setLlmModels(s.models);
        setLlmOnline(s.online);
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const saveLLM = async () => {
    try {
      await api.saveLLMSettings({ endpoint: llmEndpoint.trim(), model: llmModel.trim() });
      toast(`✅ 已保存 Ollama 配置`);
    } catch (e) {
      setErr(`保存失败：${String(e)}`);
    }
  };

  const testLLM = async () => {
    try {
      const res = await api.testLLMSettings({ endpoint: llmEndpoint.trim(), model: llmModel.trim() });
      setLlmModels(res.models);
      setLlmOnline(true);
      setMsg(`✅ 连接成功（${res.endpoint}），可用模型：${res.models.length} 个`);
      setTimeout(() => setMsg(''), 4000);
    } catch (e) {
      setLlmOnline(false);
      setErr(`连接失败：${String(e)}`);
    }
  };

  const techs = useMemo(
    () => [...new Set(exercises.map((e) => e.tech_stack))].sort(),
    [exercises],
  );
  const subcats = useMemo(
    () => categories.filter((c) => c.tech_stack === tech).sort((a, b) => a.subcategory.localeCompare(b.subcategory, 'zh-CN')),
    [categories, tech],
  );

  const rows = useMemo(() => {
    let list = exercises.filter((e) => e.tech_stack === tech);
    if (filterSub) list = list.filter((e) => e.subcategory === filterSub);
    const q = search.trim().toLowerCase();
    if (q) list = list.filter((e) => e.id.toLowerCase().includes(q) || e.title.toLowerCase().includes(q));
    return list;
  }, [exercises, tech, filterSub, search]);

  const toast = (m: string) => {
    setMsg(m);
    setErr('');
    setTimeout(() => setMsg(''), 3000);
  };

  const exportLibrary = async () => {
    try {
      const list = await api.exportExercises();
      const blob = new Blob([JSON.stringify(list, null, 2)], { type: 'application/json' });
      const a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = `codegym-exercises-${new Date().toISOString().slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(a.href);
      toast(`✅ 已导出 ${list.length} 道题`);
    } catch (e) {
      setErr(`导出失败：${String(e)}`);
    }
  };

  const onPickFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    try {
      const data = JSON.parse(await file.text());
      const arr = Array.isArray(data) ? data : [data];
      const res = await api.importExercises(arr);
      toast(`✅ 导入成功：${res.imported}/${res.total} 道题`);
      refresh().catch(() => undefined);
    } catch (err2) {
      setErr(`读取文件失败：${String(err2)}`);
    }
  };

  // --- category management ---
  const addSub = async () => {
    if (!tech || !newSub.trim()) return;
    try {
      await api.createCategory(tech, newSub.trim());
      setNewSub('');
      toast(`✅ 已创建分类「${newSub.trim()}」`);
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`创建分类失败：${String(e)}`);
    }
  };

  const renameSub = async (sub: string) => {
    const res = window.prompt(`为「${sub}」输入新名称：`, sub);
    if (!res || res === sub || !res.trim()) return;
    try {
      await api.renameCategory(tech, sub, res.trim());
      toast(`✅ 分类「${sub}」已重命名为「${res.trim()}」，其题目已一并移动`);
      if (filterSub === sub) setFilterSub('');
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`重命名失败：${String(e)}`);
    }
  };

  const removeSub = async (sub: string) => {
    if (!window.confirm(`确认删除分类「${sub}」？其下的题目将移回「基础」。`)) return;
    try {
      await api.deleteCategory(tech, sub);
      toast(`🗑️ 已删除分类「${sub}」，题目移回「基础」`);
      if (filterSub === sub) setFilterSub('');
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`删除分类失败：${String(e)}`);
    }
  };

  // --- exercise management ---
  const moveExercise = async (id: string, sub: string) => {
    try {
      const d = await api.getExercise(id);
      await api.updateExercise(id, { ...d, subcategory: sub });
      toast(`✅ 已移动 ${id} → ${sub}`);
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`移动失败：${String(e)}`);
    }
  };

  const removeExercise = async (id: string) => {
    if (!window.confirm(`确认删除题目 ${id}？`)) return;
    try {
      await api.deleteExercise(id);
      toast(`🗑️ 已删除 ${id}`);
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`删除失败：${String(e)}`);
    }
  };

  const openEdit = async (id?: string, prefill?: ExerciseDetail) => {
    try {
      if (prefill) {
        const d = { ...prefill, tech_stack: tech, subcategory: '基础' };
        setEditing(d);
      } else if (id) {
        setEditing(await api.getExercise(id));
      } else {
        setEditing({ ...EMPTY_EXERCISE, tech_stack: tech });
      }
      setJson('');
      setErr('');
    } catch (e) {
      setErr(`加载失败：${String(e)}`);
    }
  };

  const onEditJsonChange = async (j: string) => {
    setJson(j);
    if (editing) setEditing({ ...editing, ...safeParse(j) });
  };

  const saveEdit = async () => {
    if (!editing) return;
    try {
      const body = json ? (JSON.parse(json) as ExerciseDetail) : editing;
      if (!body.id) throw new Error('缺少 id 字段');
      await api.createExercise(body);
      toast(`✅ 已保存 ${body.id}`);
      setEditing(null);
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`保存失败：${String(e)}`);
    }
  };

  const runAiGenerate = async () => {
    if (!aiTech || aiBusy) return;
    setAiBusy(true);
    setAiResults([]);
    setErr('');
    setMsg('⏳ 正在生成（本地模型较慢，约 1-3 分钟），请耐心等待…');
    try {
      const res = await api.aiGenerate({
        tech_stack: aiTech,
        weight: aiWeight,
        knowledge_points: aiKnowledge,
        count: aiCount,
      });
      setAiResults(res.items);
      setMsg(`✅ 生成完成：成功 ${res.ok}/${res.total} 道`);
      refresh().catch(() => undefined);
    } catch (e) {
      setErr(`AI 生成失败：${String(e)}`);
      setMsg('');
    } finally {
      setAiBusy(false);
    }
  };

  return (
    <div className="page">
      <section className="panel">
        <div className="panel-title">📦 题库管理</div>
        <div className="tech-filter">
          {techs.map((t) => (
            <button
              key={t}
              className={`tech-btn ${tech === t ? 'active' : ''}`}
              onClick={() => {
                setTech(t);
                setFilterSub('');
              }}
            >
              {TECH_EMOJI[t] ?? '📦'} {t}（{exercises.filter((e) => e.tech_stack === t).length}）
            </button>
          ))}
          <span className="spacer" />
          <input
            className="lib-search"
            placeholder="搜索题目 ID / 标题…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button className="ghost-btn" onClick={() => void exportLibrary()}>⬇️ 导出</button>
          <button className="ghost-btn" onClick={() => fileRef.current?.click()}>📂 导入</button>
          <input ref={fileRef} type="file" accept=".json,application/json" onChange={onPickFile} hidden />
          <button className="primary-btn" onClick={() => void openEdit(undefined)}>＋ 新建题目</button>
        </div>
        {msg && <div className="ok-banner">{msg}</div>}
        {err && <div className="error-banner">⚠️ {err}</div>}
      </section>

      <section className="panel">
        <div className="panel-title">⚙️ Ollama 设置</div>
        <div className="ai-form">
          <label className="ai-endpoint">
            服务地址
            <input
              value={llmEndpoint}
              placeholder="http://localhost:11434"
              onChange={(e) => setLlmEndpoint(e.target.value)}
            />
          </label>
          <label>
            模型
            <input
              list="llm-models"
              value={llmModel}
              placeholder="qwen3.5:latest"
              onChange={(e) => setLlmModel(e.target.value)}
            />
            <datalist id="llm-models">
              {llmModels.map((m) => (
                <option key={m} value={m} />
              ))}
            </datalist>
          </label>
          <button className="ghost-btn" onClick={() => void testLLM()}>
            测试连接
          </button>
          <button className="primary-btn" onClick={() => void saveLLM()}>
            保存
          </button>
          <span className={`llm-state ${llmOnline ? 'ok' : ''}`}>
            {llmOnline ? '● 在线' : '○ 离线'}
          </span>
        </div>
      </section>

      <section className="panel">
        <div className="panel-title">🤖 AI 生成题目（本地 Ollama）</div>
        <div className="ai-form">
          <label>
            技术栈
            <select value={aiTech} onChange={(e) => setAiTech(e.target.value)}>
              {techs.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
          </label>
          <label>
            负重
            <select value={aiWeight} onChange={(e) => setAiWeight(e.target.value)}>
              {['beginner', 'easy', 'medium', 'hard', 'expert'].map((w) => (
                <option key={w} value={w}>
                  {w}
                </option>
              ))}
            </select>
          </label>
          <label>
            数量
            <input
              type="number"
              min={1}
              max={5}
              value={aiCount}
              onChange={(e) => setAiCount(Number(e.target.value))}
            />
          </label>
          <label className="ai-knowledge">
            知识点（逗号分隔）
            <input
              value={aiKnowledge}
              placeholder="如：列表推导式, 过滤偶数"
              onChange={(e) => setAiKnowledge(e.target.value)}
            />
          </label>
          <button className="primary-btn" onClick={() => void runAiGenerate()} disabled={aiBusy || !aiTech}>
            {aiBusy ? '⏳ 生成中…' : '🚀 开始生成'}
          </button>
        </div>
        {aiResults.length > 0 && (
          <div className="ai-results">
            {aiResults.map((r) => (
              <div key={r.id} className={`ai-result ${r.ok ? 'ok' : 'err'}`}>
                <span className="ai-status">{r.ok ? '✅' : '❌'}</span>
                <span className="ai-title">{r.title || r.id}</span>
                <span className="ai-id">{r.id}</span>
                {r.ok ? (
                  <button className="ghost-btn sm" onClick={() => void openEdit(r.id)}>编辑</button>
                ) : (
                  <span className="ai-err">{r.error}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </section>

      <section className="panel">
        <div className="panel-title">
          分类管理（{tech || '—'}）
          <span className="panel-hint-inline">新建分类、重命名、删除（删除后题目移回「基础」）</span>
        </div>
        <div className="subcat-grid">
          {subcats.map((c) => (
            <div key={c.subcategory} className={`subcat-chip ${filterSub === c.subcategory ? 'active' : ''}`}>
              <button className="subcat-name" onClick={() => setFilterSub(filterSub === c.subcategory ? '' : c.subcategory)}>
                {c.subcategory}
                <span className="subcat-count">{c.exercise_count}</span>
              </button>
              <button className="mini-btn" title="重命名" onClick={() => void renameSub(c.subcategory)}>✏️</button>
              <button className="mini-btn" title="删除分类" onClick={() => void removeSub(c.subcategory)}>🗑️</button>
            </div>
          ))}
          <div className="subcat-add">
            <input
              placeholder="新分类名（如：高级 / 实战 / web）"
              value={newSub}
              onChange={(e) => setNewSub(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && void addSub()}
            />
            <button className="primary-btn" onClick={() => void addSub()} disabled={!newSub.trim()}>
              ＋
            </button>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="panel-title">
          题目管理{filterSub ? ` · ${filterSub}` : ''}（{rows.length}）
        </div>
        {rows.length === 0 ? (
          <div className="empty-state">该分类下暂无题目，点右上角「＋新建题目」或导入。</div>
        ) : (
          <div className="admin-list">
            <div className="admin-row admin-head">
              <span className="admin-id">ID</span>
              <span className="admin-title">标题</span>
              <span className="admin-type">题型</span>
              <span className="admin-subw">所属分类</span>
              <span className="admin-ops">操作</span>
            </div>
            {rows.map((ex) => (
              <div key={ex.id} className="admin-row">
                <span className="admin-id">{ex.id}</span>
                <span className="admin-title">{ex.title}</span>
                <span className="admin-type">{TYPE_LABEL[ex.exercise_type] ?? ex.exercise_type}</span>
                <select
                  className="admin-subw"
                  value={ex.subcategory}
                  onChange={(e) => void moveExercise(ex.id, e.target.value)}
                >
                  {subcats.map((c) => (
                    <option key={c.subcategory} value={c.subcategory}>
                      {c.subcategory}
                    </option>
                  ))}
                </select>
                <span className="admin-ops">
                  <button className="ghost-btn sm" onClick={() => void openEdit(ex.id)}>编辑</button>
                  <button className="danger-btn" onClick={() => void removeExercise(ex.id)}>删除</button>
                </span>
              </div>
            ))}
          </div>
        )}
      </section>

      {editing && (
        <div className="modal-backdrop" onClick={() => setEditing(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="panel-title">
              ✍️ 编辑题目：{editing.id || '（新建）'}
              <button className="ghost-btn sm" onClick={() => setEditing(null)}>关闭</button>
            </div>
            <textarea
              className="json-editor"
              value={json || JSON.stringify(editing, null, 2)}
              spellCheck={false}
              onChange={(e) => onEditJsonChange(e.target.value)}
            />
            {err && <div className="error-banner">⚠️ {err}</div>}
            <div className="admin-actions">
              <button className="primary-btn" onClick={() => void saveEdit()}>保存</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function safeParse(s: string): Partial<ExerciseDetail> {
  try {
    return JSON.parse(s) as Partial<ExerciseDetail>;
  } catch {
    return {};
  }
}
