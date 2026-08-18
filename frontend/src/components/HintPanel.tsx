import { useState } from 'react';
import type { VariantFillPoint } from '../types';

interface HintPanelProps {
  fillPoints: VariantFillPoint[];
  onHintUsed: (pointId: string, level: number) => void;
  readOnly?: boolean;
}

const PENALTY = ['0 分', '-5 分', '-15 分', '-30 分'];

export default function HintPanel({ fillPoints, onHintUsed, readOnly }: HintPanelProps) {
  const [levels, setLevels] = useState<Record<string, number>>({});
  const [selected, setSelected] = useState<string>(fillPoints[0]?.id ?? '');

  const point = fillPoints.find((p) => p.id === selected) ?? fillPoints[0];
  if (!point) return <div className="hint-panel empty">本题没有可用的辅助保护</div>;

  const level = levels[point.id] ?? 0;

  const reveal = () => {
    if (readOnly || level >= 4) return;
    const next = level + 1;
    setLevels((s) => ({ ...s, [point.id]: next }));
    onHintUsed(point.id, next);
  };

  const hintTexts = [point.hint_level_1, point.hint_level_2, point.hint_level_3, point.hint_level_4];

  return (
    <div className="hint-panel">
      <div className="hint-toolbar">
        <span>💡 辅助保护</span>
        <select
          value={selected}
          disabled={readOnly}
          onChange={(e) => setSelected(e.target.value)}
        >
          {fillPoints.map((p) => (
            <option key={p.id} value={p.id}>
              训练点：{p.id}（{p.category}）
            </option>
          ))}
        </select>
        <button className="hint-btn" onClick={reveal} disabled={readOnly || level >= 4}>
          {level >= 4 ? '已显示答案' : `💡 查看提示 (Level ${level + 1})`}
        </button>
      </div>
      <div className="hint-levels">
        {hintTexts.map((text, i) => (
          <div
            key={i}
            className={`hint-row ${i < level ? 'revealed' : 'locked'}`}
            title={`Level ${i + 1}：${PENALTY[i]}`}
          >
            <span className="hint-level">L{i + 1}</span>
            <span className="hint-cost">{PENALTY[i]}</span>
            <span className="hint-text">{i < level ? text : '🔒 点击上方按钮逐级解锁'}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
