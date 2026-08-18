import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { navigate } from '../App';
import { TECH_EMOJI, TYPE_LABEL } from '../tech';
import type { ExerciseListItem, ProgressResponse } from '../types';

export default function HomePage() {
  const [exercises, setExercises] = useState<ExerciseListItem[]>([]);
  const [progress, setProgress] = useState<ProgressResponse | null>(null);
  const [tech, setTech] = useState('');
  const [favOnly, setFavOnly] = useState(false);
  const [favorites, setFavorites] = useState<Set<string>>(new Set());
  const [error, setError] = useState('');

  useEffect(() => {
    api.listExercises(tech).then(setExercises).catch((e) => setError(String(e)));
    api.progress().then(setProgress).catch(() => undefined);
  }, [tech]);

  useEffect(() => {
    api.favorites().then((list) => setFavorites(new Set(list.map((f) => f.exercise_id)))).catch(() => undefined);
  }, []);

  const toggleFav = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const res = await api.toggleFavorite(id);
    setFavorites((prev) => {
      const next = new Set(prev);
      if (res.favorited) next.add(id);
      else next.delete(id);
      return next;
    });
  };

  const p = progress?.progress;
  const completed = p ? Object.keys(p.completed_exercises).length : 0;
  const total = Math.max(exercises.length, 1);
  const pct = total ? Math.round((completed / total) * 100) : 0;
  const techs = [...new Set(exercises.map((e) => e.tech_stack))];
  const shown = favOnly ? exercises.filter((e) => favorites.has(e.id)) : exercises;

  return (
    <div className="page">
      {p && (
        <section className="dashboard">
          <div className="panel">
            <div className="panel-title">💪 今日训练</div>
            <div className="dash-grid">
              <div className="dash-item">
                <span className="dash-label">整体进度</span>
                <div className="progress-bar">
                  <div className="progress-fill" style={{ width: `${pct}%` }} />
                </div>
                <span className="dash-value">
                  {pct}% ({completed}/{total} 组)
                </span>
              </div>
              <div className="dash-item">
                <span className="dash-label">连续训练</span>
                <span className="dash-value big">🔥 {p.current_streak} 天</span>
                <span className="dash-sub">最长 {p.longest_streak} 天</span>
              </div>
              <div className="dash-item">
                <span className="dash-label">累计训练</span>
                <span className="dash-value big">{p.total_attempts} 组</span>
                <span className="dash-sub">正确 {p.correct_attempts} 组</span>
              </div>
              <div className="dash-item">
                <span className="dash-label">辅助使用</span>
                <span className="dash-value big">{p.total_hint_used} 次</span>
                <span className="dash-sub">少用辅助，记忆更牢</span>
              </div>
            </div>
          </div>
          {Object.keys(p.weak_points).length > 0 && (
            <div className="panel">
              <div className="panel-title">📌 薄弱部位</div>
              <div className="weak-chip-row">
                {Object.entries(p.weak_points)
                  .filter(([k]) => !k.endsWith('#hint'))
                  .map(([id]) => {
                    const ex = exercises.find((e) => e.id === id);
                    return (
                      <button
                        key={id}
                        className="chip"
                        onClick={() => navigate({ name: 'training', id })}
                      >
                        {ex ? ex.title : id} 🔁
                      </button>
                    );
                  })}
              </div>
            </div>
          )}
        </section>
      )}

      <section className="panel">
        <div className="panel-title">📚 训练计划</div>
        <div className="tech-filter">
          <button className={`tech-btn ${tech === '' ? 'active' : ''}`} onClick={() => setTech('')}>
            全部
          </button>
          {techs.map((t) => (
            <button
              key={t}
              className={`tech-btn ${tech === t ? 'active' : ''}`}
              onClick={() => setTech(t)}
            >
              {TECH_EMOJI[t] ?? '📦'} {t}
            </button>
          ))}
          <button
            className={`tech-btn fav-filter ${favOnly ? 'active' : ''}`}
            onClick={() => setFavOnly((v) => !v)}
          >
            ⭐ 收藏（{favorites.size}）
          </button>
        </div>

        {error && <div className="error-banner">⚠️ {error}</div>}

        {favOnly && shown.length === 0 && (
          <div className="empty-state">还没有收藏的题目，给需要加强的题目点个 ⭐ 吧。</div>
        )}

        <div className="exercise-grid">
          {shown.map((ex) => (
            <div
              key={ex.id}
              className="exercise-card"
              role="button"
              tabIndex={0}
              onClick={() => navigate({ name: 'training', id: ex.id })}
              onKeyDown={(e) => e.key === 'Enter' && navigate({ name: 'training', id: ex.id })}
            >
              <div className="card-head">
                <span className="type-badge">{TYPE_LABEL[ex.exercise_type] ?? ex.exercise_type}</span>
                <button
                  className={`fav-btn ${favorites.has(ex.id) ? 'fav-on' : ''}`}
                  title={favorites.has(ex.id) ? '取消收藏' : '收藏，有空多练练'}
                  onClick={(e) => void toggleFav(ex.id, e)}
                >
                  {favorites.has(ex.id) ? '★' : '☆'}
                </button>
              </div>
              <div className="card-title">{ex.title}</div>
              <div className="card-desc">{ex.description}</div>
              <div className="card-foot">
                <span>
                  {TECH_EMOJI[ex.tech_stack] ?? ''} {ex.tech_stack} · 训练点 {ex.point_count} 个
                </span>
                <span className="card-go">开始训练 →</span>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
