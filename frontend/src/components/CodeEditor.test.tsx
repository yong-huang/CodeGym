import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import CodeEditor, { extractAnswers, toPlaceholderText } from './CodeEditor';
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