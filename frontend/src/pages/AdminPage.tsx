import { useEffect, useState } from 'react';
import { api, adminToken } from '../api/client';
import type { AdminUser, BankSource, Environment, ExerciseDetail, ExerciseListItem } from '../types';

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
  // environments
  const [envs, setEnvs] = useState<Environment[]>([]);
  const [envName, setEnvName] = useState('');
  const [envType, setEnvType] = useState('local');
  const [envDetail, setEnvDetail] = useState('');
  // admin token（后端启用 -admin-token 时必填）
  const [tokenInput, setTokenInput] = useState(adminToken.get());
  // 用户管理
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [newUsername, setNewUsername] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [newDisplayName, setNewDisplayName] = useState('');
  // 题库动态加载
  const [bankSources, setBankSources] = useState<BankSource[]>([]);
  const [bankName, setBankName] = useState('');

  const refresh = () => {
    api.listExercises().then(setExercises).catch(() => undefined);
    api.adminStats().then(setStats).catch(() => undefined);
    api.adminEnvironments().then(setEnvs).catch(() => undefined);
    api.adminUsers().then(setUsers).catch(() => undefined);
    api.adminBanks().then(setBankSources).catch(() => undefined);
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

  const toast = (m: string) => {
    setMessage(m);
    setError('');
    setTimeout(() => setMessage(''), 3000);
  };

  const addEnvironment = async () => {
    setMessage(''); setError('');
    if (!envName.trim()) { setError('请输入环境名称'); return; }
    try {
      const cfg: Record<string, unknown> = {};
      if (envDetail.trim()) {
        try { Object.assign(cfg, JSON.parse(envDetail)); } catch { cfg.description = envDetail.trim(); }
      }
      if (envType === 'local' && !cfg.interpreter) cfg.interpreter = envName.trim();
      await api.createEnvironment({ name: envName.trim(), type: envType, config: cfg, enabled: true });
      setEnvName(''); setEnvDetail('');
      toast(`✅ 已创建环境`);
      refresh();
    } catch (e) { setError(`创建失败：${String(e)}`); }
  };

  const testEnvironment = async (id: string) => {
    try {
      const res = await api.testEnvironment(id);
      toast(`${res.health_status === 'healthy' ? '✅' : '❌'} ${id} → ${res.detail}`);
      refresh();
    } catch (e) { setError(`测试失败：${String(e)}`); }
  };

  const removeEnvironment = async (id: string) => {
    if (!window.confirm(`确认删除环境 ${id}？`)) return;
    try {
      await api.deleteEnvironment(id);
      toast(`🗑️ 已删除环境 ${id}`);
      refresh();
    } catch (e) { setError(`删除失败：${String(e)}`); }
  };

  const importLibrary = async (list?: ExerciseDetail[]) => {    setMessage('');
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

  // --- 用户管理 ---
  const createUser = async () => {
    setMessage(''); setError('');
    if (!newUsername.trim() || newPassword.length < 6) {
      setError('用户名必填，密码至少 6 位');
      return;
    }
    try {
      await api.adminCreateUser({
        username: newUsername.trim(),
        password: newPassword,
        display_name: newDisplayName.trim(),
      });
      setNewUsername(''); setNewPassword(''); setNewDisplayName('');
      toast(`✅ 已创建用户`);
      refresh();
    } catch (e) { setError(`创建失败：${String(e)}`); }
  };

  const toggleUser = async (u: AdminUser) => {
    setMessage(''); setError('');
    try {
      await api.adminUpdateUser(u.id, { disabled: !u.disabled });
      toast(u.disabled ? `✅ 已启用 ${u.username}` : `🚫 已禁用 ${u.username}`);
      refresh();
    } catch (e) { setError(`操作失败：${String(e)}`); }
  };

  const resetPassword = async (u: AdminUser) => {
    const pwd = window.prompt(`为 ${u.username} 设置新密码（至少 6 位）`);
    if (!pwd) return;
    if (pwd.length < 6) { setError('密码至少 6 位'); return; }
    try {
      await api.adminUpdateUser(u.id, { password: pwd });
      toast(`🔑 已重置 ${u.username} 的密码`);
    } catch (e) { setError(`重置失败：${String(e)}`); }
  };

  const removeUser = async (u: AdminUser) => {
    if (!window.confirm(`确认删除用户 ${u.username}？其进度/收藏/作答记录将一并删除。`)) return;
    try {
      await api.adminDeleteUser(u.id);
      toast(`🗑️ 已删除用户 ${u.username}`);
      refresh();
    } catch (e) { setError(`删除失败：${String(e)}`); }
  };

  // --- 题库动态加载 ---
  const uploadBank = async (file: File) => {
    setMessage(''); setError('');
    const name = (bankName.trim() || file.name.replace(/\.json$/i, '')).replace(/[^A-Za-z0-9_-]/g, '-');
    try {
      const content = await file.text();
      const res = await api.adminUploadBank({ name, content });
      toast(`✅ 题库 ${res.name} 已加载 ${res.imported} 题，立即可练`);
      setBankName('');
      refresh();
    } catch (e) { setError(`上传失败：${String(e)}`); }
  };

  const onPickBankFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (file) void uploadBank(file);
  };

  return (
    <div className="page">
      <section className="panel">
        <div className="panel-title">🔐 管理令牌</div>
        <p className="panel-hint">
          后端以 `-admin-token`（或环境变量 CODEGYM_ADMIN_TOKEN）启动时，教练端接口需要令牌。
          本地未启用鉴权时可留空。
        </p>
        <div className="env-add">
          <input
            type="password"
            placeholder="X-Admin-Token（仅存本浏览器）"
            value={tokenInput}
            onChange={(e) => setTokenInput(e.target.value)}
          />
          <button
            className="primary-btn"
            onClick={() => {
              adminToken.set(tokenInput);
              toast(tokenInput ? '✅ 令牌已保存，后续教练端请求将自动携带' : '🗑️ 已清除令牌');
            }}
          >
            保存
          </button>
        </div>
      </section>

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
        <div className="panel-title">👥 用户管理</div>
        <p className="panel-hint">
          每个用户的进度 / 收藏 / 薄弱点独立保存；用户可在「我的题库」选择练习范围。禁用后无法登录。
        </p>
        <div className="admin-list">
          {users.map((u) => (
            <div key={u.id} className="admin-row">
              <span className="admin-id">{u.username}</span>
              <span className="admin-title">{u.display_name || '-'}</span>
              <span className="admin-sub">{u.total_attempts} 次作答</span>
              <span className="admin-sub">完成 {u.completed_count}</span>
              <span className="admin-sub">🔥 {u.current_streak} 天</span>
              <span className={`admin-type ${u.disabled ? 'user-disabled' : ''}`}>
                {u.disabled ? '已禁用' : '正常'}
              </span>
              <button className="ghost-btn sm" onClick={() => void toggleUser(u)}>
                {u.disabled ? '启用' : '禁用'}
              </button>
              <button className="ghost-btn sm" onClick={() => void resetPassword(u)}>
                重置密码
              </button>
              <button className="danger-btn" onClick={() => void removeUser(u)}>
                删除
              </button>
            </div>
          ))}
          {users.length === 0 && <div className="empty-state">暂无用户</div>}
          <div className="env-add">
            <input placeholder="用户名" value={newUsername} onChange={(e) => setNewUsername(e.target.value)} />
            <input
              type="password"
              placeholder="密码（至少 6 位）"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
            />
            <input placeholder="昵称（可空）" value={newDisplayName} onChange={(e) => setNewDisplayName(e.target.value)} />
            <button className="primary-btn" onClick={() => void createUser()}>
              ＋ 添加用户
            </button>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="panel-title">📥 题库动态加载</div>
        <p className="panel-hint">
          上传题库 JSON（与 data/bank 同格式）立即生效，无需重启；也可以直接把 JSON 文件放进 data/bank
          目录，后端每 15 秒自动扫描热加载（可用 `-bank-watch` 调整间隔）。
        </p>
        <div className="admin-list">
          {bankSources.map((b) => (
            <div key={`${b.source}:${b.name}`} className="admin-row">
              <span className="admin-id">{b.name}</span>
              <span className="admin-type">{b.source === 'upload' ? '上传' : '目录'}</span>
              <span className="admin-sub">{b.exercise_count} 题</span>
              <span className="admin-sub">{new Date(b.updated_at).toLocaleString()}</span>
            </div>
          ))}
          {bankSources.length === 0 && <div className="empty-state">还没有动态加载的题库</div>}
          <div className="env-add">
            <input
              placeholder="题库名（可空，默认取文件名）"
              value={bankName}
              onChange={(e) => setBankName(e.target.value)}
            />
            <label className="file-btn primary-btn">
              📂 选择题库 JSON 上传
              <input type="file" accept=".json,application/json" onChange={onPickBankFile} hidden />
            </label>
          </div>
        </div>
      </section>

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
        <div className="panel-title">⚙️ 执行环境管理</div>
        <div className="env-list">
          {envs.map((e) => (
            <div key={e.id} className="env-row">
              <span className={`env-status ${e.health_status === 'healthy' ? 'ok' : ''}`}>{e.health_status === 'healthy' ? '✅' : '❔'}</span>
              <span className="env-name">{e.name}</span>
              <span className="env-type">{e.type}</span>
              <span className="env-state">{e.enabled ? '启用' : '停用'}</span>
              <button className="ghost-btn sm" onClick={() => void testEnvironment(e.id)}>测试</button>
              <button className="danger-btn" onClick={() => void removeEnvironment(e.id)}>删除</button>
            </div>
          ))}
          <div className="env-add">
            <input placeholder="环境名（如 node-20）" value={envName} onChange={(e) => setEnvName(e.target.value)} />
            <select value={envType} onChange={(e) => setEnvType(e.target.value)}>
              <option value="local">local</option>
              <option value="docker">docker</option>
              <option value="kubernetes">kubernetes</option>
              <option value="pyodide">pyodide</option>
            </select>
            <input placeholder="config（JSON，可空）" value={envDetail} onChange={(e) => setEnvDetail(e.target.value)} />
            <button className="primary-btn" onClick={() => void addEnvironment()}>＋ 添加</button>
          </div>
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
