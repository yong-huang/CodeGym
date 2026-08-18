import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import Sidebar from './Sidebar';

describe('Sidebar', () => {
  it('renders the library header and empty tree at SSR', () => {
    const html = renderToStaticMarkup(
      <Sidebar collapsed={false} activeExerciseId={undefined} />,
    );
    expect(html).toContain('📚 题库');
    expect(html).toContain('sidebar-tree');
  });

  it('collapses to an empty rail (toggle lives in the topbar)', () => {
    const collapsed = renderToStaticMarkup(
      <Sidebar collapsed activeExerciseId={undefined} />,
    );
    expect(collapsed).toContain('sidebar collapsed');
    // the collapsed rail must not overflow with tree content
    expect(collapsed).not.toContain('sidebar-tree');
  });
});