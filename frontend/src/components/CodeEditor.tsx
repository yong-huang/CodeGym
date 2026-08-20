import { useMemo, useRef } from 'react';
import type { VariantFillPoint } from '../types';

interface CodeEditorProps {
  /** Original template containing ____FILL_<id>____ tokens (fill mode). */
  markedCode?: string;
  fillPoints: VariantFillPoint[];
  /** Raw editable text (controlled from the parent). */
  value: string;
  onChange: (value: string) => void;
  readOnly?: boolean;
}

interface Segment {
  type: 'code' | 'fill';
  text?: string;
  id?: string;
}

export interface FillRange {
  id: string;
  start: number;
  end: number;
}

const TOKEN = /____FILL_([A-Za-z0-9_]+)____/g;

/** Placeholder shown inside every blank. It lives in the real text (so the
 *  transparent textarea and the highlight layer stay perfectly aligned) and
 *  is auto-selected on focus/click so typing replaces it instantly. */
const PLACEHOLDER = '____';

// --- template parsing / answer extraction -------------------------------

function parseTemplate(markedCode: string, fillIds: string[]): Segment[] {
  const idSet = new Set(fillIds);
  const segs: Segment[] = [];
  let rest = markedCode;
  TOKEN.lastIndex = 0;
  for (;;) {
    const m = TOKEN.exec(rest);
    if (!m) {
      if (rest) segs.push({ type: 'code', text: rest });
      break;
    }
    if (m.index > 0) segs.push({ type: 'code', text: rest.slice(0, m.index) });
    const id = m[1];
    segs.push(idSet.has(id) ? { type: 'fill', id } : { type: 'code', text: m[0] });
    rest = rest.slice(m.index + m[0].length);
    TOKEN.lastIndex = 0;
  }
  return segs;
}

/** Replace fill tokens with the placeholder so the editor starts fillable.
 *  The placeholder is auto-selected on click, so typing replaces it directly. */
export function toPlaceholderText(markedCode: string): string {
  return markedCode.replace(TOKEN, PLACEHOLDER);
}

/**
 * Locate the code segments (the fixed text outside the fills) inside the live
 * edited text via backtracking so the whole text is covered exactly: each fill
 * answer is whatever sits between two adjacent code segments. A plain
 * first-occurrence anchor would mis-place a fill when its text already
 * contains the following code segment (e.g. answer `.upper()` and the next
 * code segment is `)`), truncating the answer. Backtracking over all segment
 * occurrences and requiring exact coverage resolves that ambiguity.
 */
function locateCodeSegments(codeTexts: string[], text: string, allowTrailing: boolean): number[] | null {
  const res: number[] = [];
  const place = (i: number, from: number): boolean => {
    if (i === codeTexts.length) {
      if (allowTrailing) return true; // leftover behind the last code segment is the trailing fill
      return from === text.length; // otherwise the last segment must end exactly at the end
    }
    const s = codeTexts[i];
    if (!s) {
      res.push(from);
      if (place(i + 1, from)) return true;
      res.pop();
      return false;
    }
    let idx = text.indexOf(s, from);
    while (idx >= 0) {
      res.push(idx);
      if (place(i + 1, idx + s.length)) return true;
      res.pop();
      idx = text.indexOf(s, idx + 1);
    }
    return false;
  };
  if (!place(0, 0)) return null;
  return res;
}

/** Greedy first-occurrence extraction; kept as a fallback for text where the
 *  user edited code outside the fills (exact coverage then fails). */
function greedyExtract(
  segs: Segment[],
  text: string,
): { answers: Record<string, string>; ranges: FillRange[] } {
  const answers: Record<string, string> = {};
  const ranges: FillRange[] = [];

  const codePos: number[] = new Array(segs.length).fill(-1);
  let search = 0;
  for (let i = 0; i < segs.length; i++) {
    const s = segs[i];
    if (s.type !== 'code' || !s.text) continue;
    let idx = text.indexOf(s.text, search);
    if (idx < 0) idx = text.indexOf(s.text); // tolerate edits before this anchor
    codePos[i] = idx;
    if (idx >= 0) search = idx + s.text.length;
  }

  let prevEnd = 0;
  for (let i = 0; i < segs.length; i++) {
    const s = segs[i];
    if (s.type === 'code') {
      if (codePos[i] >= 0 && s.text) prevEnd = codePos[i] + s.text.length;
      continue;
    }
    let nextStart = text.length;
    for (let j = i + 1; j < segs.length; j++) {
      if (segs[j].type === 'code' && codePos[j] >= 0) {
        nextStart = codePos[j];
        break;
      }
    }
    const start = Math.min(prevEnd, text.length);
    const end = Math.max(nextStart, start);
    ranges.push({ id: s.id!, start, end });
    answers[s.id!] = text.slice(start, end);
  }
  return { answers, ranges };
}

/**
 * Locate the fill regions inside the *live* edited text by anchoring on the
 * surrounding (unchanged) code segments, then extract each fill's answer.
 */
export function extractAnswers(
  markedCode: string,
  text: string,
  fillIds: string[],
): { answers: Record<string, string>; ranges: FillRange[] } {
  const segs = parseTemplate(markedCode, fillIds);
  const answers: Record<string, string> = {};
  const ranges: FillRange[] = [];

  const codeSegs: { text: string; segIdx: number }[] = [];
  for (let i = 0; i < segs.length; i++) {
    const s = segs[i];
    if (s.type === 'code' && s.text) codeSegs.push({ text: s.text, segIdx: i });
  }
  const hasTrailingFill = segs[segs.length - 1]?.type === 'fill';

  const positions = locateCodeSegments(codeSegs.map((c) => c.text), text, hasTrailingFill);
  if (!positions) return greedyExtract(segs, text);

  const posBySeg = new Map<number, number>();
  codeSegs.forEach((seg, i) => posBySeg.set(seg.segIdx, positions[i]));

  let prevEnd = 0;
  for (let i = 0; i < segs.length; i++) {
    const s = segs[i];
    if (s.type === 'code') {
      const p = posBySeg.get(i);
      if (p !== undefined && s.text) prevEnd = p + s.text.length;
      continue;
    }
    let nextStart = text.length;
    for (let j = i + 1; j < segs.length; j++) {
      if (segs[j].type === 'code') {
        const p = posBySeg.get(j);
        if (p !== undefined) {
          nextStart = p;
          break;
        }
      }
    }
    const start = Math.min(prevEnd, text.length);
    const end = Math.max(nextStart, start);
    ranges.push({ id: s.id!, start, end });
    answers[s.id!] = text.slice(start, end);
  }
  return { answers, ranges };
}

// --- rendering -----------------------------------------------------------

const INLINE_STYLE: React.CSSProperties = {
  fontFamily: "'SF Mono', 'JetBrains Mono', Menlo, Consolas, monospace",
  fontSize: 13,
  lineHeight: 1.6,
  tabSize: 4,
  letterSpacing: 0,
  whiteSpace: 'pre',
};

function buildHighlight(text: string, ranges: FillRange[]) {
  // return per-line nodes with fill ranges wrapped in underline spans.
  // Zero-width (empty) fills render as an inline "type here" blank marker.
  const nodes: React.ReactNode[] = [];
  let lineStart = 0;
  for (let line = 0; ; line++) {
    const nl = text.indexOf('\n', lineStart);
    const lineEnd = nl === -1 ? text.length : nl;
    const parts: React.ReactNode[] = [];
    let cursor = lineStart;

    // markers on this line, sorted by position
    const markers = ranges
      .filter((r) => {
        if (r.start === r.end) return r.start >= lineStart && r.start <= lineEnd;
        return r.start < lineEnd && r.end > lineStart;
      })
      .sort((a, b) => a.start - b.start || a.end - b.end);

    for (const r of markers) {
      if (r.start === r.end) {
        if (r.start > cursor) parts.push(<span key={`c${cursor}`}>{text.slice(cursor, r.start)}</span>);
        parts.push(<span key={`fe${r.id}@${r.start}`} className="fill-underline fill-empty" />);
        cursor = r.start;
        continue;
      }
      const s = Math.max(r.start, lineStart);
      const e = Math.min(r.end, lineEnd);
      if (s > cursor) parts.push(<span key={`c${cursor}`}>{text.slice(cursor, s)}</span>);
      const seg = text.slice(s, e);
      const cls = seg === PLACEHOLDER ? 'fill-underline fill-placeholder' : 'fill-underline';
      parts.push(
        <span key={`f${r.id}@${s}`} className={cls}>
          {seg}
        </span>,
      );
      cursor = e;
    }
    if (cursor < lineEnd) parts.push(<span key={`c${cursor}`}>{text.slice(cursor, lineEnd)}</span>);
    // empty source lines must keep their height or the overlay drifts from the
    // textarea line by line (click/typing would land on the wrong line).
    if (parts.length === 0) {
      parts.push(
        <span key={`e${line}`} className="empty-line">
          {'\u00A0'}
        </span>,
      );
    }
    nodes.push(
      <div key={line} className="code-line">
        {parts}
      </div>,
    );
    if (nl === -1) break;
    lineStart = nl + 1;
  }
  return nodes;
}

export default function CodeEditor({
  markedCode = '',
  fillPoints,
  value,
  onChange,
  readOnly,
}: CodeEditorProps) {
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const fillIds = useMemo(() => fillPoints.map((p) => p.id), [fillPoints]);

  const ranges = useMemo(
    () => (markedCode ? extractAnswers(markedCode, value, fillIds).ranges : []),
    [markedCode, value, fillIds],
  );

  const lines = useMemo(() => value.split('\n').length, [value]);
  const highlight = useMemo(() => buildHighlight(value, ranges), [value, ranges]);

  const selectOptions = useMemo(
    () =>
      fillPoints.filter((p) => p.fill_type === 'select' && p.options?.length).flatMap((p) => p.options ?? []),
    [fillPoints],
  );

  // Clicking into an untouched blank auto-selects its placeholder so the
  // first keystroke replaces it — no manual deleting of underscores.
  const selectPlaceholderAtCaret = () => {
    const ta = inputRef.current;
    if (!ta || readOnly) return;
    if (ta.selectionStart === null || ta.selectionStart !== ta.selectionEnd) return;
    const pos = ta.selectionStart;
    const blank = ranges.find(
      (r) =>
        r.start <= pos &&
        pos <= r.end &&
        r.end - r.start > 0 &&
        value.slice(r.start, r.end) === PLACEHOLDER,
    );
    if (blank) ta.setSelectionRange(blank.start, blank.end);
  };

  const gutterLines = useMemo(() => {
    const out: React.ReactNode[] = [];
    for (let i = 1; i <= lines; i++) out.push(<div key={i}>{i}</div>);
    return out;
  }, [lines]);

  return (
    <div className="editor-shell">
      {selectOptions.length > 0 && !readOnly && (
        <div className="editor-options">
          可选项：{selectOptions.map((o) => <code key={o} className="chip-opt">{o}</code>)}
        </div>
      )}
      {/* This element is the scroll container: the textarea (in-flow, defines
          height) and the highlight/gutter (absolute) are all its children, so
          they scroll together perfectly — click/caret alignment is automatic. */}
      <div className="editor-body">
        <div className="editor-gutter" style={INLINE_STYLE}>
          {gutterLines}
        </div>
        <pre className="editor-pre" style={INLINE_STYLE}>
          {highlight}
        </pre>
        <textarea
          ref={inputRef}
          className="editor-input"
          style={INLINE_STYLE}
          wrap="off"
          spellCheck={false}
          rows={Math.max(lines, 2)}
          value={value}
          readOnly={readOnly}
          aria-label="代码编辑区"
          onChange={(e) => onChange(e.target.value)}
          onFocus={selectPlaceholderAtCaret}
          onMouseUp={selectPlaceholderAtCaret}
          onClick={selectPlaceholderAtCaret}
        />
      </div>
    </div>
  );
}
