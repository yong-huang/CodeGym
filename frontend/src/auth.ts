import { useEffect, useState } from 'react';
import { api, authToken, UNAUTHORIZED_EVENT } from './api/client';
import type { User } from './types';

// Module-level auth state shared by every page. `loaded` distinguishes
// "not fetched yet" from "fetched, not logged in".
let user: User | null = null;
let loaded = false;
const listeners = new Set<() => void>();

function emit() {
  listeners.forEach((fn) => fn());
}

/** Fetches the current user (no token -> logged out). */
export async function refreshUser(): Promise<User | null> {
  if (!authToken.get()) {
    user = null;
    loaded = true;
    emit();
    return null;
  }
  try {
    const res = await api.me();
    user = res.user;
  } catch {
    user = null;
  }
  loaded = true;
  emit();
  return user;
}

/** Applies a fresh user object (e.g. after saving bank selection). */
export function setUser(u: User | null) {
  user = u;
  loaded = true;
  emit();
}

/** Logs out locally (token cleared) without a server call. */
export function clearAuth() {
  authToken.clear();
  user = null;
  loaded = true;
  emit();
}

/** Revokes the session server-side and clears local state. */
export async function logout() {
  try {
    await api.logout();
  } catch {
    // token may already be invalid; still clear locally
  }
  clearAuth();
}

if (typeof window !== 'undefined') {
  window.addEventListener(UNAUTHORIZED_EVENT, () => {
    user = null;
    loaded = true;
    emit();
  });
}

function subscribe(fn: () => void) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

/** React hook returning the current user (null when logged out). */
export function useUser(): User | null {
  const [u, setU] = useState<User | null>(user);
  useEffect(() => {
    setU(user);
    return subscribe(() => setU(user));
  }, []);
  return u;
}

/** React hook reporting whether the initial auth check finished. */
export function useAuthReady(): boolean {
  const [ready, setReady] = useState(loaded);
  useEffect(() => {
    setReady(loaded);
    return subscribe(() => setReady(loaded));
  }, []);
  return ready;
}

/** Selected "tech::sub" keys of the current user (empty = all banks). */
export function selectedBanksOf(u: User | null): Set<string> {
  return new Set(u?.selected_banks ?? []);
}
