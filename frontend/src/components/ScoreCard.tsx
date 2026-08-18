import type { ExecuteResult } from '../types';

interface ScoreCardProps {
  result: ExecuteResult;
}

const GRADE_META: Record<string, { label: string; cls: string }> = {
  S: { label: '🎉 完美！完全掌握', cls: 'grade-s' },
  A: { label: '✅ 优秀，继续加油', cls: 'grade-a' },
  B: { label: '📖 良好，建议复习薄弱点', cls: 'grade-b' },
  C: { label: '📝 需要加强训练', cls: 'grade-c' },
  D: { label: '🔄 建议重新学习基础知识', cls: 'grade-d' },
};

export default function ScoreCard({ result }: ScoreCardProps) {
  const meta = GRADE_META[result.grade] ?? GRADE_META.D;
  const passRate = result.fill_results.length
    ? Math.round(
        (result.fill_results.filter((r) => r.is_correct).length / result.fill_results.length) * 100,
      )
    : 0;

  return (
    <div className="score-card">
      <div className={`score-head ${meta.cls}`}>
        <div className="score-main">
          <span className="score-num">{Math.round(result.score)}</span>
          <span className="score-unit">分</span>
          <span className="score-grade">{result.grade}</span>
        </div>
        <div className="score-meta">
          <div className="score-status">{result.is_correct ? '✅ 已通关' : '⚠️ 未通关'}</div>
          <div className="score-label">{meta.label}</div>
          <div className="score-sub">
            通过率 {passRate}% · 用时 {result.execution_time_ms} ms
          </div>
        </div>
      </div>

      {result.logs.length > 0 && (
        <div className="score-logs">
          {result.logs.map((log, i) => (
            <div key={i} className="log-line">
              {log}
            </div>
          ))}
        </div>
      )}

      <div className="score-output">
        <div className="panel-title">▶ 执行输出</div>
        <pre className="output-box">{result.output || '（无输出）'}</pre>
        {result.expected_output && (
          <>
            <div className="panel-title">🎯 预期输出</div>
            <pre className="output-box expected">{result.expected_output}</pre>
          </>
        )}
      </div>

      {result.fill_results.length > 0 && (
        <div className="score-detail">
          <div className="panel-title">📊 详细评分</div>
          <table className="fill-table">
            <thead>
              <tr>
                <th>训练点</th>
                <th>你的答案</th>
                <th>正确值</th>
                <th>得分</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {result.fill_results.map((r) => (
                <tr key={r.id}>
                  <td>{r.point_id}</td>
                  <td>
                    <code className="answer-cell">{truncate(r.user_answer, 40)}</code>
                  </td>
                  <td>
                    <code className="answer-cell">{truncate(r.expected, 40)}</code>
                  </td>
                  <td>{Math.round(r.score)}</td>
                  <td>{r.is_correct ? '✅ 正确' : '❌ 错误'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {result.weak_analysis && (
        <div className="score-analysis">
          <div className="panel-title">💡 薄弱分析</div>
          <p>{result.weak_analysis}</p>
        </div>
      )}

      {(result.correct_answer || result.explanation) && (
        <div className="score-answer">
          <div className="panel-title">✅ 正确答案 / 📖 解析</div>
          {result.correct_answer && (
            <pre className="output-box answer-code">{result.correct_answer}</pre>
          )}
          {result.explanation && <p className="answer-expl">{result.explanation}</p>}
        </div>
      )}
    </div>
  );
}

function truncate(s: string, n: number) {
  const r = s.split(/\s+/).join(' ');
  return r.length > n ? r.slice(0, n) + '…' : r;
}
