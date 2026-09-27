import { useState } from 'react';
import { api, authToken } from '../api/client';
import { refreshUser } from '../auth';
import { navigate } from '../App';

function errorText(e: unknown): string {
  return String(e).replace(/^Error:\s*/, '');
}

export default function LoginPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      const res =
        mode === 'login'
          ? await api.login({ username, password })
          : await api.register({ username, password, display_name: displayName });
      authToken.set(res.token);
      await refreshUser();
      navigate({ name: 'home' });
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="auth-wrap">
      <div className="auth-card">
        <div className="auth-logo">🏋️</div>
        <h1 className="auth-title">CodeGym</h1>
        <p className="auth-sub">练出来的，才是自己的</p>
        <div className="auth-tabs">
          <button
            type="button"
            className={`auth-tab ${mode === 'login' ? 'active' : ''}`}
            onClick={() => setMode('login')}
          >
            登录
          </button>
          <button
            type="button"
            className={`auth-tab ${mode === 'register' ? 'active' : ''}`}
            onClick={() => setMode('register')}
          >
            注册
          </button>
        </div>
        <form className="auth-form" onSubmit={submit}>
          <label className="auth-field">
            <span>用户名</span>
            <input
              value={username}
              autoComplete="username"
              placeholder="3-32 位字母/数字/下划线"
              onChange={(e) => setUsername(e.target.value)}
            />
          </label>
          {mode === 'register' && (
            <label className="auth-field">
              <span>昵称（可选）</span>
              <input
                value={displayName}
                placeholder="显示在顶栏的名字"
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </label>
          )}
          <label className="auth-field">
            <span>密码</span>
            <input
              type="password"
              value={password}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              placeholder="至少 6 位"
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          {error && <div className="error-banner">⚠️ {error}</div>}
          <button className="auth-submit" type="submit" disabled={busy}>
            {busy ? '请稍候…' : mode === 'login' ? '登录' : '注册并进入'}
          </button>
        </form>
        <p className="auth-hint">
          {mode === 'login' ? '还没有账号？点上方「注册」创建，数据独立保存。' : '每个账号的进度、收藏、薄弱点相互独立。'}
        </p>
      </div>
    </div>
  );
}
