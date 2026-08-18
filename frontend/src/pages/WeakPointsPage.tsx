import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { navigate } from '../App';
import type { WeakPointItem } from '../types';

export default function WeakPointsPage() {
  const [items, setItems] = useState<WeakPointItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .weakPoints()
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="page">
      <section className="panel">
        <div className="panel-title">📝 薄弱部位（错题本）</div>
        <p className="panel-hint">多次出错的训练动作会记录在这里，建议针对性加练。</p>
        {loading ? (
          <div className="loading">加载中…</div>
        ) : items.length === 0 ? (
          <div className="empty-state">🎉 太棒了，目前没有薄弱部位，继续加油！</div>
        ) : (
          <div className="weak-list">
            {items.map((item) => (
              <button
                key={item.exercise_id}
                className="weak-row"
                onClick={() => navigate({ name: 'training', id: item.exercise_id })}
              >
                <span className="weak-fail">❌ ×{item.fail_count}</span>
                <span className="weak-title">{item.title}</span>
                <span className="weak-go">重新加练 →</span>
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
