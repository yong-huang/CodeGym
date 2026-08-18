import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { ExerciseDetail, ExerciseListItem } from '../types';

const EMPTY_EXERCISE: ExerciseDetail = {
  id: '',
  title: '',
  description: '',
  exercise_type: 'fill',
  tech_stack: 'python',
  weight: 'medium',
  category: '',
  subcategory: '基础',
  full_code: '',
  fill_points: { points: [] },
  spotter_hints: { level_1: '', level_2: '', level_3: '', level_4: '' },
  time_limit_seconds: 30,
  enabled: true,
};

export default function AdminPage() {
  const [exercises, setExercises] = useState<ExerciseListItem[]>([]);
  const [stats, setStats] = useState<Record<string, unknown>>({});
  const [json, setJson] = useState(JSON.stringify(EMPTY_EXERCISE, null, 2));
  const [importText, setImportText] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const refresh = () => {
    api.listExercises().then(setExercises).catch(() => undefined);
    api.adminStats().then(setStats).catch(() => undefined);
  };

  useEffect(refresh, []);

  const submit = async () => {
    setMessage('');
    setError('');
    try {
      const ex = JSON.parse(json) as ExerciseDetail;
      if (ex.id) {
        await api.createExercise(ex);
        setMessage(`✅ 已创建训练动作 ${ex.id}`);
      } else {
        setError('缺少 id 字段');
      }
      refresh();
    } catch (e) {
      setError(`保存失败：${String(e)}`);
    }
  };

  const remove = async (id: string) => {
    if (!window.confirm(`确认删除训练动作 ${id}？`)) return;
    await api.deleteExercise(id);
    refresh();
  };

  const exportLibrary = async () => {
    setMessage('');
    setError('');
    try {
      const list = await api.exportExercises();
      const blob = new Blob([JSON.stringify(list, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `codegym-exercises-${new Date().toISOString().slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(url);
      setMessage(`✅ 已导出 ${list.length} 道题`);
    } catch (e) {
      setError(`导出失败：${String(e)}`);
    }
  };

  const importLibrary = async (list?: ExerciseDetail[]) => {
    setMessage('');
    setError('');
    try {
      const data = list ?? (JSON.parse(importText) as unknown);
      const arr = Array.isArray(data) ? data : [data]; // 容错：单个对象也允许
      if (!Array.isArray(arr) || arr.length === 0) {
        throw new Error('没有可导入的题目（需要数组或单个对象）');
      }
      const res = await api.importExercises(arr as ExerciseDetail[]);
      setMessage(`✅ 导入成功：${res.imported}/${res.total} 道题`);
      refresh();
    } catch (e) {
      setError(`导入失败：${String(e)}`);
    }
  };

  const onPickFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = ''; // 允许重复选择同一文件
    if (!file) return;
    try {
      const text = await file.text();
      setImportText(text);
      await importLibrary(JSON.parse(text) as ExerciseDetail[]);
    } catch (err) {
      setError(`读取文件失败：${String(err)}`);
    }
  };

  return (
    <div className="page">
      {Object.keys(stats).length > 0 && (
        <section className="dashboard">
          <div className="panel">
            <div className="panel-title">📊 平台统计</div>
            <div className="dash-grid">
              <div className="dash-item">
                <span className="dash-label">训练动作</span>
                <span className="dash-value big">{String(stats.exercise_count ?? 0)}</span>
              </div>
              <div className="dash-item">
                <span className="dash-label">总作答</span>
                <span className="dash-value big">{String(stats.attempt_count ?? 0)}</span>
              </div>
              <div className="dash-item">
                <span className="dash-label">训练学员</span>
                <span className="dash-value big">{String(stats.user_count ?? 0)}</span>
              </div>
              <div className="dash-item">
                <span className="dash-label">人均作答</span>
                <span className="dash-value big">{String(stats.avg_attempts_per_user ?? 0)}</span>
              </div>
            </div>
          </div>
        </section>
      )}

      <section className="panel">
        <div className="panel-title">📋 训练动作管理</div>
        <div className="admin-list">
          {exercises.map((ex) => (
            <div key={ex.id} className="admin-row">
              <span className="admin-id">{ex.id}</span>
              <span className="admin-title">{ex.title}</span>
              <span className="admin-tech">{ex.tech_stack}</span>
              <span className="admin-sub">{ex.subcategory}</span>
              <span className="admin-type">{ex.exercise_type}</span>
              <button className="danger-btn" onClick={() => void remove(ex.id)}>
                删除
              </button>
            </div>
          ))}
        </div>
      </section>

      <section className="panel">
        <div className="panel-title">✍️ 创建 / 更新训练动作（JSON）</div>
        <p className="panel-hint">
          教练端 JSON 编辑。可直接粘贴 GET /api/v1/exercises/:id 返回的完整对象后修改。
        </p>
        <textarea className="json-editor" value={json} spellCheck={false} onChange={(e) => setJson(e.target.value)} />
        {message && <div className="ok-banner">{message}</div>}
        {error && <div className="error-banner">⚠️ {error}</div>}
        <div className="admin-actions">
          <button className="primary-btn" onClick={() => void submit()}>
            保存训练动作
          </button>
        </div>
      </section>

      <section className="panel">
        <div className="panel-title">📦 题库导入 / 导出</div>
        <p className="panel-hint">
          导出：将整个题库（全部技术栈）下载为一个 JSON 文件，可迁移 / 备份 / 共享给其他 CodeGym 实例。
        </p>
        <div className="admin-actions">
          <button className="primary-btn" onClick={() => void exportLibrary()}>
            ⬇️ 导出题库
          </button>
        </div>

        <p className="panel-hint" style={{ marginTop: 16 }}>
          导入（推荐）：直接选择导出的 .json 文件，按 id 覆盖或新增。
        </p>
        <label className="file-btn primary-btn">
          📂 选择文件导入
          <input type="file" accept=".json,application/json" onChange={onPickFile} hidden />
        </label>

        <p className="panel-hint" style={{ marginTop: 16 }}>
          或粘贴 JSON（支持数组或单个对象）：
        </p>
        <textarea
          className="json-editor import-editor"
          value={importText}
          spellCheck={false}
          placeholder='粘贴题目的 JSON，例如 [{"id":"py-lc-001",...}, ...]'
          onChange={(e) => setImportText(e.target.value)}
        />
        {message && <div className="ok-banner">{message}</div>}
        {error && <div className="error-banner">⚠️ {error}</div>}
        <div className="admin-actions">
          <button className="primary-btn" onClick={() => void importLibrary()} disabled={!importText.trim()}>
            ⬆️ 导入题库
          </button>
        </div>
      </section>
    </div>
  );
}
