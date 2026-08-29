import { useCallback, useEffect, useState } from 'react';
import { api } from '../api/client';
import CodeEditor, { toPlaceholderText } from '../components/CodeEditor';
import HintPanel from '../components/HintPanel';
import ScoreCard from '../components/ScoreCard';
import WeightSelector, { WEIGHTS } from '../components/WeightSelector';
import type { ExecuteResult, Variant, WeightLevel } from '../types';

const TYPE_LABEL: Record<string, string> = {
  fill: '填空补全',
  bugfix: 'Bug 修复',
  command: '命令补全',
  config: '配置补全',
};

export default function TrainingPage({ id }: { id: string }) {
  const [variant, setVariant] = useState<Variant | null>(null);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [code, setCode] = useState('');
  const [hintUsage, setHintUsage] = useState<Record<string, number>>({});
  const [favorited, setFavorited] = useState(false);
  const [result, setResult] = useState<ExecuteResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState('');

  const loadVariant = useCallback(
    async (levelToLoad: WeightLevel) => {
      setLoading(true);
      setResult(null);
      setHintUsage({});
      setError('');
      try {
        const v = await api.getVariant(id, levelToLoad);
        setVariant(v);
        setAnswers({});
        // bugfix exercises are fully editable; fill exercises start with blanks
        setCode(v.exercise_type === 'bugfix' ? v.code_for_edit ?? '' : toPlaceholderText(v.marked_code));
      } catch (e) {
        setError(String(e));
      } finally {
        setLoading(false);
      }
    },
    [id],
  );

  useEffect(() => {
    void loadVariant('medium');
  }, [loadVariant]);

  useEffect(() => {
    api.favorites().then((list) => setFavorited(list.some((f) => f.exercise_id === id))).catch(() => undefined);
  }, [id]);

  const toggleFav = async () => {
    const res = await api.toggleFavorite(id);
    setFavorited(res.favorited);
  };

  const changeWeight = (w: WeightLevel) => {
    void loadVariant(w);
  };

  const onCodeChange = (next: string) => {
    setCode(next);
  };

  const onAnswersChange = useCallback((ans: Record<string, string>) => {
    setAnswers(ans);
  }, []);

  const onHint = (pointId: string, levelNo: number) => {
    setHintUsage((s) => ({ ...s, [pointId]: Math.max(s[pointId] ?? 0, levelNo) }));
  };

  const run = async () => {
    if (!variant) return;
    setRunning(true);
    setError('');
    try {
      const payload: Record<string, unknown> = {
        exercise_id: variant.exercise_id,
        variant_id: variant.variant_id,
        weight_level: variant.weight_level,
        fill_answers: Object.fromEntries(
          Object.entries(answers).map(([id, value]) => [`FILL_${id}`, value]),
        ),
        hint_usage: hintUsage,
      };
      if (variant.exercise_type === 'bugfix') {
        payload.user_code = code;
      }
      const res = await api.execute(payload);
      setResult(res);
    } catch (e) {
      setError(String(e));
    } finally {
      setRunning(false);
    }
  };

  if (loading && !variant) {
    return <div className="page loading">加载训练动作…</div>;
  }
  if (!variant) {
    return <div className="page">⚠️ {error || '未找到该训练动作'}</div>;
  }

  const weightMeta = WEIGHTS.find((w) => w.key === variant.weight_level);

  return (
    <div className="page">
      <section className="panel exercise-header">
        <div className="ex-head-main">
          <span className="type-badge">{TYPE_LABEL[variant.exercise_type]}</span>
          <h2>🏋️ {variant.title}</h2>
          <button
            className={`fav-btn ${favorited ? 'fav-on' : ''}`}
            title={favorited ? '取消收藏' : '收藏，有空多练练'}
            onClick={() => void toggleFav()}
          >
            {favorited ? '★ 已收藏' : '☆ 收藏'}
          </button>
        </div>
        <div className="ex-desc">{variant.description}</div>
        <div className="ex-meta">
          <span>技术栈：{variant.tech_stack}</span>
          <span>
            负重：{weightMeta?.label}（{weightMeta?.desc}）
          </span>
          <span>训练点：{variant.fill_points?.length ?? 0} 个</span>
        </div>
        {variant.bug_description && (
          <div className="bug-banner">
            <strong>🐞 Bug 描述：</strong>
            {variant.bug_description}
          </div>
        )}
      </section>

      <section className="panel">
        <div className="panel-title">🎯 选择负重</div>
        <WeightSelector value={variant.weight_level} onChange={changeWeight} disabled={!!result} />
      </section>

      {error && <div className="error-banner">⚠️ {error}</div>}

      <section className="panel">
        <div className="panel-title">
          🧑‍💻 代码编辑区
          {result && <span className="locked-tag">已提交，本次训练锁定</span>}
        </div>
        <p className="panel-hint">
          {variant.exercise_type === 'bugfix'
            ? '整段代码可直接编辑，修复 Bug 后点击运行验证。'
            : '整段代码均可编辑，虚线处为训练点，把内容填进虚线区域即可。'}
        </p>
        <CodeEditor
          markedCode={variant.exercise_type === 'bugfix' ? undefined : variant.marked_code}
          fillPoints={variant.exercise_type === 'bugfix' ? [] : (variant.fill_points ?? [])}
          value={code}
          onChange={onCodeChange}
          onAnswers={onAnswersChange}
          readOnly={!!result}
        />
      </section>

      <section className="panel">
        <div className="panel-title">💡 辅助保护（点击逐级解锁，使用会扣分）</div>
        <HintPanel key={variant.variant_id} fillPoints={variant.fill_points ?? []} onHintUsed={onHint} readOnly={!!result} />
      </section>

      <section className="panel actions">
        <button className="primary-btn" onClick={run} disabled={running || !!result}>
          {running ? '⏳ 运行中…' : '▶ 运行验证'}
        </button>
        <button className="ghost-btn" onClick={() => void loadVariant(variant.weight_level)}>
          🔄 换负重再练
        </button>
      </section>

      {result && <ScoreCard result={result} />}
    </div>
  );
}
