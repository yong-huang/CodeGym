import type { WeightLevel } from '../types';

interface WeightSelectorProps {
  value: WeightLevel;
  onChange: (level: WeightLevel) => void;
  disabled?: boolean;
}

export const WEIGHTS: { key: WeightLevel; label: string; desc: string }[] = [
  { key: 'beginner', label: '🌱 入门', desc: '1-2 个训练点' },
  { key: 'easy', label: '🌿 简单', desc: '2-3 个训练点' },
  { key: 'medium', label: '🌳 中等', desc: '3-4 个训练点' },
  { key: 'hard', label: '🔥 困难', desc: '4-6 个训练点' },
  { key: 'expert', label: '💎 专家', desc: '6+ 个训练点' },
];

export default function WeightSelector({ value, onChange, disabled }: WeightSelectorProps) {
  return (
    <div className="weight-selector">
      {WEIGHTS.map((w) => (
        <button
          key={w.key}
          className={`weight-btn ${value === w.key ? 'active' : ''}`}
          disabled={disabled}
          onClick={() => onChange(w.key)}
        >
          <span className="weight-label">{w.label}</span>
          <span className="weight-desc">{w.desc}</span>
        </button>
      ))}
    </div>
  );
}
