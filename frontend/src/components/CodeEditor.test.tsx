import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import CodeEditor, { extractAnswers, toPlaceholderText, updateRanges } from './CodeEditor';
import type { VariantFillPoint } from '../types';

const TEMPLATE =
  'numbers = [1, 2, 3]\nresult = [____FILL_expr____ for x in ____FILL_iter____ if x % 2 == 0]\nprint(result)';

const FILL_POINTS: VariantFillPoint[] = [
  { id: 'expr', category: 'expression', fill_type: 'text', line: 2, hint_level_1: '', hint_level_2: '', hint_level_3: '', hint_level_4: '' },
  { id: 'iter', category: 'iterator', fill_type: 'text', line: 2, hint_level_1: '', hint_level_2: '', hint_level_3: '', hint_level_4: '' },
];

describe('toPlaceholderText', () => {
  it('replaces every fill token with the placeholder blank', () => {
    const text = toPlaceholderText(TEMPLATE);
    expect(text).not.toContain('____FILL_');
    expect(text).toContain('__'); // the visible blank the user clicks into
    expect(text).toBe(
      'numbers = [1, 2, 3]\nresult = [____ for x in ____ if x % 2 == 0]\nprint(result)',
    );
  });
});

describe('extractAnswers', () => {
  it('extracts answers typed into the blanks (single line)', () => {
    const edited = TEMPLATE.replace('____FILL_expr____', 'x**2').replace('____FILL_iter____', 'numbers');
    const { answers } = extractAnswers(TEMPLATE, edited, ['expr', 'iter']);
    expect(answers.expr).toBe('x**2');
    expect(answers.iter).toBe('numbers');
  });

  it('returns ranges that line up with the fills', () => {
    const edited = TEMPLATE.replace('____FILL_expr____', 'x**2').replace('____FILL_iter____', 'numbers');
    const { ranges } = extractAnswers(TEMPLATE, edited, ['expr', 'iter']);
    expect(ranges).toHaveLength(2);
    expect(edited.slice(ranges[0].start, ranges[0].end)).toBe('x**2');
    expect(edited.slice(ranges[1].start, ranges[1].end)).toBe('numbers');
    expect(ranges[0].start < ranges[1].start).toBe(true);
  });

  it('treats an untouched blank as the placeholder (auto-selected, not graded as correct)', () => {
    const { answers } = extractAnswers(TEMPLATE, toPlaceholderText(TEMPLATE), ['expr', 'iter']);
    expect(answers.expr).toBe('____');
    expect(answers.iter).toBe('____');
  });

  it('typing over a blank extracts the typed text (replaces the placeholder)', () => {
    const template = 'result = [____FILL_expr____ for x in numbers if x % 2 == 0]';
    const edited = 'result = [x**2 for x in numbers if x % 2 == 0]';
    const { answers, ranges } = extractAnswers(template, edited, ['expr']);
    expect(answers.expr).toBe('x**2');
    expect(edited.slice(ranges[0].start, ranges[0].end)).toBe('x**2');
  });

  it('handles fills at the very start and end of the code', () => {
    const t = '____FILL_a____ + ____FILL_b____';
    const edited = 'x + y';
    const { answers, ranges } = extractAnswers(t, edited, ['a', 'b']);
    expect(answers.a).toBe('x');
    expect(answers.b).toBe('y');
    expect(edited.slice(ranges[0].start, ranges[0].end)).toBe('x');
    expect(edited.slice(ranges[1].start, ranges[1].end)).toBe('y');
  });

  it('tolerates whitespace differences in surrounding code', () => {
    const edited = TEMPLATE.replace('____FILL_expr____', 'x * * 2') // different spacing
      .replace('____FILL_iter____', 'numbers');
    const { answers } = extractAnswers(TEMPLATE, edited, ['expr', 'iter']);
    expect(answers.expr).toBe('x * * 2');
    expect(answers.iter).toBe('numbers');
  });

  it('extracts a multi-line code answer', () => {
    const template = 'def f():\n  ____FILL_body____\n  return 0';
    const edited = 'def f():\n  x = 1\n  y = 2\n  return x + y\n  z = 9\n  return 0';
    const { answers } = extractAnswers(template, edited, ['body']);
    expect(answers.body).toContain('x = 1');
    expect(answers.body).toContain('y = 2');
  });

  it('ignores fill ids with no token in the template', () => {
    const { answers, ranges } = extractAnswers(TEMPLATE, TEMPLATE, ['expr', 'nonexistent']);
    expect(answers.expr).toBeDefined();
    expect(answers.nonexistent).toBeUndefined();
    expect(ranges).toHaveLength(1);
    const filled = TEMPLATE.replace('____FILL_expr____', 'x**2');
    const { answers: a2 } = extractAnswers(TEMPLATE, filled, ['expr', 'nonexistent']);
    expect(a2.expr).toBe('x**2');
  });

  it('keeps the full answer when it contains the following code segment text', () => {
    // regression: `("go".upper())` — the answer `.upper()` already contains a
    // `)` that is also the next code segment; it must not be truncated to `.upper(`.
    const template = 'print("go"____FILL_upper____)';
    const edited = 'print("go".upper())';
    const { answers, ranges } = extractAnswers(template, edited, ['upper']);
    expect(answers.upper).toBe('.upper()');
    expect(edited.slice(ranges[0].start, ranges[0].end)).toBe('.upper()');
  });

  it('still reports a genuinely incomplete answer as-is', () => {
    const template = 'print("go"____FILL_upper____)';
    const edited = 'print("go".upper()'; // user really typed a missing `)`
    const { answers } = extractAnswers(template, edited, ['upper']);
    expect(answers.upper).toBe('.upper(');
  });
});

describe('CodeEditor rendering', () => {
  it('renders an editable textarea with line numbers and underlined fills', () => {
    const value = toPlaceholderText(TEMPLATE);
    const html = renderToStaticMarkup(
      <CodeEditor markedCode={TEMPLATE} fillPoints={FILL_POINTS} value={value} onChange={() => {}} />,
    );
    // editable textarea holding the placeholder text (SSR renders value as children)
    expect(html).toContain('<textarea');
    expect(html).toContain('>numbers = [1, 2, 3]');
    // line numbers gutter for the 3 source lines
    expect(html).toContain('editor-gutter');
    expect(html).toContain('<div>1</div>');
    expect(html).toContain('<div>3</div>');
    // each blank is a dashed-underline span
    expect(html.match(/fill-underline/g)?.length).toBe(2);
    // never leaks the raw answer token into the editable value
    expect(html).not.toContain('____FILL_');
  });

  it('keeps underlines aligned with where the user typed', () => {
    const value = TEMPLATE.replace('____FILL_expr____', 'x**2').replace('____FILL_iter____', 'numbers');
    const html = renderToStaticMarkup(
      <CodeEditor markedCode={TEMPLATE} fillPoints={FILL_POINTS} value={value} onChange={() => {}} />,
    );
    // answers now appear as underlined spans
    expect(html).toContain('fill-underline');
    expect(html).toContain('x**2');
    expect(html).toContain('numbers');
  });

  it('places blanks between code segments, not at the line start', () => {
    const template =
      'const result = numbers.filter(x => ____FILL_filter_cond____).map(x => ____FILL_map_fn____);\nconsole.log(result);';
    const points: VariantFillPoint[] = [
      { ...FILL_POINTS[0], id: 'filter_cond', category: 'condition' },
      { ...FILL_POINTS[1], id: 'map_fn', category: 'expression' },
    ];
    const html = renderToStaticMarkup(
      <CodeEditor
        markedCode={template}
        fillPoints={points}
        value={toPlaceholderText(template)}
        onChange={() => {}}
      />,
    );
    const filterIdx = html.indexOf('filter(x =&gt; ');
    const firstBlank = html.indexOf('fill-underline fill-placeholder');
    const mapIdx = html.indexOf(').map(x =&gt; ');
    const secondBlank = html.lastIndexOf('fill-underline fill-placeholder');
    expect(filterIdx).toBeGreaterThan(0);
    expect(filterIdx).toBeLessThan(firstBlank); // blank sits right after filter(x =>
    expect(firstBlank).toBeLessThan(mapIdx); // then the .map chain continues
    expect(mapIdx).toBeLessThan(secondBlank); // second blank sits after .map(x =>
    // the line text must not be dumped again after the blanks (inside the <pre>)
    const preEnd = html.indexOf('</pre>');
    expect(html.slice(firstBlank, preEnd).indexOf('const result')).toBe(-1);
    expect(html.slice(secondBlank, preEnd)).toContain(');</span>');
    // the editable textarea holds the SAME placeholders (alignment guarantee)
    expect(html).toContain('numbers.filter(x =&gt; ____).map(x =&gt; ____);');
  });

  it('keeps the Go palindrome template intact after typing answers', () => {
    // regression: typing into blanks must never corrupt surrounding code
    const template =
      'package main\n\nimport "fmt"\n\nfunc isPalindrome(s string) bool {\n\trunes := []rune(s)\n\tfor ____FILL_init____; i < len(runes); ____FILL_incr____ {\n\t\tif ____FILL_cond____ {\n\t\t\treturn false\n\t\t}\n\t}\n\treturn true\n}\n\nfunc main() {\n\tfmt.Println(isPalindrome("racecar"))\n\tfmt.Println(isPalindrome("hello"))\n}';
    const typed = template
      .replace('____FILL_init____', 'i := 0')
      .replace('____FILL_incr____', 'i++')
      .replace('____FILL_cond____', 'runes[i] != runes[len(runes)-1-i]');
    const { answers } = extractAnswers(template, typed, ['init', 'incr', 'cond']);
    expect(answers.init).toBe('i := 0');
    expect(answers.incr).toBe('i++');
    expect(answers.cond).toBe('runes[i] != runes[len(runes)-1-i]');
    // the value/display text must be byte-for-byte the typed text
    expect(typed).toContain('func isPalindrome(s string) bool {');
    expect(typed).not.toContain('isPalnindrome');
    // and rendering highlights exactly three fill spans
    const html = renderToStaticMarkup(
      <CodeEditor markedCode={template} fillPoints={[
        { ...FILL_POINTS[0], id: 'init', category: 'init' },
        { ...FILL_POINTS[1], id: 'incr', category: 'iteration' },
        { ...FILL_POINTS[0], id: 'cond', category: 'condition' },
      ]} value={typed} onChange={() => {}} />,
    );
    expect(html.match(/fill-underline/g)?.length).toBe(3);
    expect(html).toContain('func isPalindrome(s string) bool {');
  });

  it('keeps empty lines tall so the overlay does not drift from the textarea', () => {
    // two blank lines before the code -> they must render as full-height rows
    const template = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println(____FILL_expr____)\n}';
    const html = renderToStaticMarkup(
      <CodeEditor
        markedCode={template}
        fillPoints={[{ ...FILL_POINTS[0], id: 'expr', category: 'expression' }]}
        value={toPlaceholderText(template)}
        onChange={() => {}}
      />,
    );
    // every source line (including blanks) is a non-empty code-line row
    const lineDivs = html.match(/class="code-line"/g);
    expect(lineDivs?.length).toBe(7); // 5 real + 2 blank lines
    expect(html).toContain('empty-line'); // blanks carry a height-preserving marker
  });
});
describe('updateRanges (增量跟踪，修复空格分隔相邻填空)', () => {
  // devops-ansible-002: `ansible webservers -m command -a "uptime"`
  const MARKED002 = 'ansible webservers ____FILL_mod_cmd____ ____FILL_args____';

  it('填第一个含空格的答案不被截断', () => {
    const initial = toPlaceholderText(MARKED002);
    const initRanges = extractAnswers(MARKED002, initial, ['mod_cmd', 'args']).ranges;
    // 占位符定位
    expect(initial.slice(initRanges[0].start, initRanges[0].end)).toBe('____');
    expect(initial.slice(initRanges[1].start, initRanges[1].end)).toBe('____');

    // 用户在第一个空输入 `-m command`（替换占位符）
    const typed = initial.replace('____', '-m command');
    const ranges = updateRanges(initial, typed, initRanges);
    const a = typed.slice(ranges[0].start, ranges[0].end);
    const b = typed.slice(ranges[1].start, ranges[1].end);
    expect(a).toBe('-m command'); // 关键：不能被截断成 `-m`
    expect(b).toBe('____');
  });

  it('再填第二个含空格的答案', () => {
    const initial = toPlaceholderText(MARKED002);
    const initRanges = extractAnswers(MARKED002, initial, ['mod_cmd', 'args']).ranges;
    const after1 = initial.replace('____', '-m command');
    const r1 = updateRanges(initial, after1, initRanges);
    const after2 = after1.replace('____', '-a "uptime"');
    const r2 = updateRanges(after1, after2, r1);
    const a = after2.slice(r2[0].start, r2[0].end);
    const b = after2.slice(r2[1].start, r2[1].end);
    expect(a).toBe('-m command');
    expect(b).toBe('-a "uptime"');
  });

  it('sql-basics-017: ORDER BY price 与 ASC 相邻填空', () => {
    const MARKED = 'SELECT name, price FROM products ____FILL_order_by____ ____FILL_asc_kw____;';
    const initial = toPlaceholderText(MARKED);
    const initRanges = extractAnswers(MARKED, initial, ['order_by', 'asc_kw']).ranges;
    const after1 = initial.replace('____', 'ORDER BY price');
    const r1 = updateRanges(initial, after1, initRanges);
    const after2 = after1.replace('____', 'ASC');
    const r2 = updateRanges(after1, after2, r1);
    expect(after2.slice(r2[0].start, r2[0].end)).toBe('ORDER BY price');
    expect(after2.slice(r2[1].start, r2[1].end)).toBe('ASC');
  });

  it('退格删除填空末尾字符时区间收缩', () => {
    const initial = toPlaceholderText(MARKED002);
    const initRanges = extractAnswers(MARKED002, initial, ['mod_cmd', 'args']).ranges;
    const typed = initial.replace('____', '-m command');
    const r1 = updateRanges(initial, typed, initRanges);
    // 删除填空内容最后一个字符 'd'（退格）
    const last = r1[0].end;
    const trimmed = typed.slice(0, last - 1) + typed.slice(last);
    const r2 = updateRanges(typed, trimmed, r1, last - 1);
    expect(trimmed.slice(r2[0].start, r2[0].end)).toBe('-m comman');
    // 第二个填空（未动的占位符）位置整体前移一位
    expect(trimmed.slice(r2[1].start, r2[1].end)).toBe('____');
  });

  it('逐字符输入含空格的答案时填空区间持续增长', () => {
    const initial = toPlaceholderText(MARKED002);
    let ranges = extractAnswers(MARKED002, initial, ['mod_cmd', 'args']).ranges;
    let text = initial;
    let prevText = initial;
    for (let i = 0; i < '-m command'.length; i++) {
      const ch = '-m command'[i];
      // 首个字符替换占位符，之后在填空末尾插入
      text =
        i === 0
          ? text.slice(0, ranges[0].start) + ch + text.slice(ranges[0].end)
          : text.slice(0, ranges[0].end) + ch + text.slice(ranges[0].end);
      const caret = ranges[0].start + i + 1;
      ranges = updateRanges(prevText, text, ranges, caret);
      prevText = text;
    }
    expect(text.slice(ranges[0].start, ranges[0].end)).toBe('-m command');
    expect(text.slice(ranges[1].start, ranges[1].end)).toBe('____');
  });
});
