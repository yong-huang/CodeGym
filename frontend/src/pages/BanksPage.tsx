import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { refreshUser, useUser } from '../auth';
import { TECH_EMOJI } from '../tech';
import type { BankInfo } from '../types';

const keyOf = (b: BankInfo) => `${b.tech_stack}::${b.subcategory}`;

export default function BanksPage() {
  const user = useUser();
  const [available, setAvailable] = useState<BankInfo[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api
      .myBanks()
      .then((res) => {
        setAvailable(res.available);
        setSelected(new Set(res.selected));
      })
      .catch((e) => setError(String(e)));
  }, []);

  const groups = useMemo(() => {
    const byTech = new Map<string, BankInfo[]>();
    for (const b of available) {
      const list = byTech.get(b.tech_stack) ?? [];
      list.push(b);
      byTech.set(b.tech_stack, list);
    }
    return [...byTech.entries()].sort((a, b) => a[0].localeCompare(b[0]));
  }, [available]);

  const totalCount = available.reduce((acc, b) => acc + b.exercise_count, 0);
  const selectedCount = available
    .filter((b) => selected.has(keyOf(b)))
    .reduce((acc, b) => acc + b.exercise_count, 0);

  const toggle = (key: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
    setMessage('');
  };

  const toggleTech = (banks: BankInfo[], on: boolean) => {
    setSelected((prev) => {
      const next = new Set(prev);
      for (const b of banks) {
        if (on) next.add(keyOf(b));
        else next.delete(keyOf(b));
      }
      return next;
    });
    setMessage('');
  };

  const save = async () => {
    setBusy(true);
    setError('');
    try {
      const res = await api.saveMyBanks([...selected]);
      setSelected(new Set(res.selected));
      await refreshUser();
      setMessage(
        res.selected.length === 0
          ? '✅ 已保存：不限制范围，练习全部题库'
          : `✅ 已保存 ${res.selected.length} 个子库，首页与侧边栏只展示所选范围`,
      );
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="page">
      <section className="panel">
        <div className="panel-title">🧪 我的题库</div>
        <p className="banks-hint">
          勾选要练习的题库范围（可按技术栈整选）；<strong>不勾选 = 练习全部题库</strong>。
          保存后首页与侧边栏只展示所选范围，进度、收藏按账号独立保存。
        </p>
        <div className="banks-actions">
          <button className="tech-btn" onClick={() => setSelected(new Set(available.map(keyOf)))}>
            全选
          </button>
          <button className="tech-btn" onClick={() => setSelected(new Set())}>
            清空
          </button>
          <span className="banks-count">
            已选 {selectedCount} 题 / 共 {totalCount} 题
          </span>
          <button className="auth-submit banks-save" onClick={save} disabled={busy}>
            {busy ? '保存中…' : '保存范围'}
          </button>
        </div>
        {message && <div className="banks-ok">{message}</div>}
        {error && <div className="error-banner">⚠️ {error}</div>}

        <div className="banks-groups">
          {groups.map(([tech, banks]) => {
            const keys = banks.map(keyOf);
            const allOn = keys.every((k) => selected.has(k));
            const someOn = keys.some((k) => selected.has(k));
            return (
              <div key={tech} className="banks-group">
                <label className="banks-tech">
                  <input
                    type="checkbox"
                    checked={allOn}
                    ref={(el) => {
                      if (el) el.indeterminate = someOn && !allOn;
                    }}
                    onChange={(e) => toggleTech(banks, e.target.checked)}
                  />
                  <span className="banks-tech-name">
                    {TECH_EMOJI[tech] ?? '📦'} {tech}
                  </span>
                  <span className="banks-tech-count">{banks.reduce((a, b) => a + b.exercise_count, 0)} 题</span>
                </label>
                <div className="banks-subs">
                  {banks.map((b) => (
                    <label key={keyOf(b)} className="banks-sub">
                      <input type="checkbox" checked={selected.has(keyOf(b))} onChange={() => toggle(keyOf(b))} />
                      <span>{b.subcategory}</span>
                      <span className="banks-sub-count">{b.exercise_count}</span>
                    </label>
                  ))}
                </div>
              </div>
            );
          })}
          {groups.length === 0 && !error && <div className="empty-state">暂无题库</div>}
        </div>
        {user && (
          <p className="banks-user">
            当前账号：<strong>{user.display_name || user.username}</strong>
          </p>
        )}
      </section>
    </div>
  );
}
