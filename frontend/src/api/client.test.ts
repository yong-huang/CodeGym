import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, authToken, UNAUTHORIZED_EVENT } from './client';

function stubBrowser() {
  const store = new Map<string, string>();
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
  });
  const listeners = new Map<string, Set<() => void>>();
  vi.stubGlobal('window', {
    addEventListener: (type: string, fn: () => void) => {
      const set = listeners.get(type) ?? new Set();
      set.add(fn);
      listeners.set(type, set);
    },
    removeEventListener: () => undefined,
    dispatchEvent: (e: Event) => {
      listeners.get(e.type)?.forEach((fn) => fn());
      return true;
    },
    location: { hash: '' },
  });
}

describe('api client auth', () => {
  beforeEach(stubBrowser);
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('stores and clears the auth token', () => {
    authToken.set('tok-1');
    expect(authToken.get()).toBe('tok-1');
    authToken.clear();
    expect(authToken.get()).toBe('');
  });

  it('sends Authorization: Bearer for logged-in requests', async () => {
    authToken.set('tok-2');
    let headers: Record<string, string> = {};
    vi.stubGlobal(
      'fetch',
      vi.fn(async (_url: string, init: RequestInit) => {
        headers = init.headers as Record<string, string>;
        return new Response(JSON.stringify({ user: { id: 'u1' } }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }),
    );
    const res = await api.me();
    expect(res.user.id).toBe('u1');
    expect(headers['Authorization']).toBe('Bearer tok-2');
  });

  it('clears the token and emits an event on 401', async () => {
    authToken.set('tok-3');
    let hit = 0;
    window.addEventListener(UNAUTHORIZED_EVENT, () => {
      hit += 1;
    });
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ error: '未登录或登录已过期' }), { status: 401 })),
    );
    await expect(api.progress()).rejects.toThrow('未登录或登录已过期');
    expect(authToken.get()).toBe('');
    expect(hit).toBe(1);
  });

  it('does not treat a failed login as a session expiry', async () => {
    authToken.set('tok-4');
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ error: '用户名或密码错误' }), { status: 401 })),
    );
    await expect(api.login({ username: 'a', password: 'b' })).rejects.toThrow('用户名或密码错误');
    expect(authToken.get()).toBe('tok-4');
  });
});
