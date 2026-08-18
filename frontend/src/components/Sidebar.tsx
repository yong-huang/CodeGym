import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { navigate } from '../App';
import { TECH_EMOJI, TYPE_LABEL } from '../tech';
import type { ExerciseListItem } from '../types';

interface SidebarProps {
  collapsed: boolean;
  activeExerciseId?: string;
}

interface SubGroup {
  subcategory: string;
  items: ExerciseListItem[];
}

interface TechGroup {
  tech: string;
  subgroups: SubGroup[];
}

const subKey = (tech: string, sub: string) => `${tech}::${sub}`;

export default function Sidebar({ collapsed, activeExerciseId }: SidebarProps) {
  const [exercises, setExercises] = useState<ExerciseListItem[]>([]);
  const [completed, setCompleted] = useState<Set<string>>(new Set());
  const [favoriteIds, setFavoriteIds] = useState<Set<string>>(new Set());
  const [favOnly, setFavOnly] = useState(false);
  const [expanded, setExpanded] = useState<Set<string>>(new Set()); // tech groups
  const [expandedSub, setExpandedSub] = useState<Set<string>>(new Set()); // subcategories

  useEffect(() => {
    const load = () => {
      api
        .listExercises()
        .then((list) => setExercises(list))
        .catch(() => undefined);
      api
        .progress()
        .then((p) => setCompleted(new Set(Object.keys(p.progress.completed_exercises))))
        .catch(() => undefined);
      api
        .favorites()
        .then((list) => setFavoriteIds(new Set(list.map((f) => f.exercise_id))))
        .catch(() => undefined);
    };
    load();
    // refresh completion marks when the user navigates back from training
    window.addEventListener('hashchange', load);
    return () => window.removeEventListener('hashchange', load);
  }, []);

  const isFav = (id: string) => favoriteIds.has(id);

  const groups = useMemo(() => {
    const byTech = new Map<string, Map<string, ExerciseListItem[]>>();
    for (const ex of exercises) {
      if (favOnly && !isFav(ex.id)) continue; // 收藏模式只展示收藏
      const subs = byTech.get(ex.tech_stack) ?? new Map<string, ExerciseListItem[]>();
      const list = subs.get(ex.subcategory) ?? [];
      list.push(ex);
      subs.set(ex.subcategory, list);
      byTech.set(ex.tech_stack, subs);
    }
    const out: TechGroup[] = [];
    for (const [tech, subs] of byTech) {
      const subgroups: SubGroup[] = [];
      for (const [sub, items] of subs) {
        subgroups.push({ subcategory: sub, items: items.sort((a, b) => a.title.localeCompare(b.title, 'zh-CN')) });
      }
      subgroups.sort((a, b) => a.subcategory.localeCompare(b.subcategory, 'zh-CN'));
      out.push({ tech, subgroups });
    }
    return out.sort((a, b) => a.tech.localeCompare(b.tech));
  }, [exercises, favOnly, favoriteIds]);

  // Groups/subcategories default to COLLAPSED. When training an exercise,
  // auto-expand its tech group and its subcategory so it stays visible.
  useEffect(() => {
    if (!activeExerciseId) return;
    const ex = exercises.find((e) => e.id === activeExerciseId);
    if (!ex) return;
    setExpanded((prev) => new Set(prev).add(ex.tech_stack));
    setExpandedSub((prev) => new Set(prev).add(subKey(ex.tech_stack, ex.subcategory)));
  }, [activeExerciseId, exercises]);

  const toggle = (setter: React.Dispatch<React.SetStateAction<Set<string>>>, key: string) => {
    setter((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const doneCount = (items: ExerciseListItem[]) => items.filter((it) => completed.has(it.id)).length;

  if (collapsed) {
    return <aside className="sidebar collapsed" aria-hidden="true" />;
  }

  return (
    <aside className="sidebar">
      <div className="sidebar-head">
        <div className="sidebar-head-row">
          <span className="sidebar-title">{favOnly ? '⭐ 收藏' : '📚 题库'}</span>
          <button
            className={`fav-filter-btn ${favOnly ? 'on' : ''}`}
            onClick={() => setFavOnly((v) => !v)}
            title={favOnly ? '显示全部题库' : '只显示收藏的题目'}
          >
            {favOnly ? '★ 全部' : '☆ 收藏'}
          </button>
        </div>
        <span className="sidebar-summary">
          {favOnly ? `${favoriteIds.size} 题收藏` : `${exercises.length} 题 · ${groups.length} 个技术栈`}
        </span>
      </div>
      <div className="sidebar-tree">
        {groups.length === 0 && <div className="sidebar-empty">暂无题目</div>}
        {groups.map((g) => {
          const open = expanded.has(g.tech);
          const activeInGroup = g.subgroups.some((sg) => sg.items.some((it) => it.id === activeExerciseId));
          const total = g.subgroups.reduce((acc, sg) => acc + sg.items.length, 0);
          const done = g.subgroups.reduce((acc, sg) => acc + doneCount(sg.items), 0);
          return (
            <div key={g.tech} className="tree-node">
              <button className={`tree-group ${activeInGroup ? 'active' : ''}`} onClick={() => toggle(setExpanded, g.tech)}>
                <span className="tree-chevron">{open ? '▾' : '▸'}</span>
                <span className="tree-emoji">{TECH_EMOJI[g.tech] ?? '📦'}</span>
                <span className="tree-group-name">{g.tech}</span>
                <span className="tree-count">
                  {done}/{total}
                </span>
              </button>
              {open && (
                <div className="tree-children">
                  {g.subgroups.map((sg) => {
                    const subOpen = expandedSub.has(subKey(g.tech, sg.subcategory));
                    const activeInSub = sg.items.some((it) => it.id === activeExerciseId);
                    return (
                      <div key={sg.subcategory} className="tree-subnode">
                        <button
                          className={`tree-subgroup ${subOpen || activeInSub ? 'active' : ''}`}
                          onClick={() => toggle(setExpandedSub, subKey(g.tech, sg.subcategory))}
                        >
                          <span className="tree-chevron">{subOpen ? '▾' : '▸'}</span>
                          <span className="tree-sub-name">{sg.subcategory}</span>
                          <span className="tree-count">
                            {doneCount(sg.items)}/{sg.items.length}
                          </span>
                        </button>
                        {subOpen && (
                          <div className="tree-children sub">
                            {sg.items.map((it) => (
                              <button
                                key={it.id}
                                className={`tree-leaf ${it.id === activeExerciseId ? 'active' : ''}`}
                                onClick={() => navigate({ name: 'training', id: it.id })}
                              >
                                <span className="tree-check">{completed.has(it.id) ? '✅' : '◻️'}</span>
                                <span className="tree-type">{TYPE_LABEL[it.exercise_type] ?? it.exercise_type}</span>
                                <span className="tree-leaf-title">{it.title}</span>
                              </button>
                            ))}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </aside>
  );
}
