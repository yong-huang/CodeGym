import type {
  AdminUser,
  AuthResponse,
  BankInfo,
  BankSource,
  CategoryItem,
  Environment,
  ExerciseDetail,
  ExerciseListItem,
  ExecuteResult,
  FavoriteItem,
  ProgressResponse,
  User,
  Variant,
  WeakPointItem,
} from '../types';
const BASE = '/api/v1';

export interface AiGenResponse {
  total: number;
  ok: number;
  items: { id: string; title: string; ok: boolean; error?: string }[];
}

export interface LLMSettings {
  endpoint: string;
  model: string;
  models: string[];
  online: boolean;
}

export const ADMIN_TOKEN_KEY = 'cg_admin_token';

export const adminToken = {
  get: () => localStorage.getItem(ADMIN_TOKEN_KEY) ?? '',
  set: (t: string) => (t ? localStorage.setItem(ADMIN_TOKEN_KEY, t.trim()) : localStorage.removeItem(ADMIN_TOKEN_KEY)),
};

export const AUTH_TOKEN_KEY = 'cg_auth_token';

export const authToken = {
  get: () => localStorage.getItem(AUTH_TOKEN_KEY) ?? '',
  set: (t: string) => (t ? localStorage.setItem(AUTH_TOKEN_KEY, t.trim()) : localStorage.removeItem(AUTH_TOKEN_KEY)),
  clear: () => localStorage.removeItem(AUTH_TOKEN_KEY),
};

/** Fired when the backend rejects the session token (401). */
export const UNAUTHORIZED_EVENT = 'codegym:unauthorized';

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((init?.headers as Record<string, string>) ?? {}),
  };
  if (authToken.get()) {
    headers['Authorization'] = `Bearer ${authToken.get()}`;
  }
  if (url.startsWith('/admin') && adminToken.get()) {
    headers['X-Admin-Token'] = adminToken.get();
  }
  const res = await fetch(`${BASE}${url}`, { ...init, headers });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    if (res.status === 401 && !url.startsWith('/auth/')) {
      authToken.clear();
      window.dispatchEvent(new Event(UNAUTHORIZED_EVENT));
    }
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  listExercises: (tech?: string) =>
    request<ExerciseListItem[]>(`/exercises${tech ? `?tech_stack=${tech}` : ''}`),

  getExercise: (id: string) => request<ExerciseDetail>(`/exercises/${id}`),

  getVariant: (id: string, level: string) =>
    request<Variant>(`/exercises/${id}/variant?level=${encodeURIComponent(level)}`),

  execute: (payload: Record<string, unknown>) =>
    request<ExecuteResult>('/execute', { method: 'POST', body: JSON.stringify(payload) }),

  progress: () => request<ProgressResponse>('/progress'),

  weakPoints: () => request<WeakPointItem[]>('/weak-points'),

  favorites: () => request<FavoriteItem[]>('/favorites'),

  toggleFavorite: (exerciseId: string) =>
    request<{ exercise_id: string; favorited: boolean }>('/favorites', {
      method: 'POST',
      body: JSON.stringify({ exercise_id: exerciseId }),
    }),

  environments: () => request<unknown[]>('/environments'),

  // --- 用户体系 ---
  register: (payload: { username: string; password: string; display_name?: string }) =>
    request<AuthResponse>('/auth/register', { method: 'POST', body: JSON.stringify(payload) }),

  login: (payload: { username: string; password: string }) =>
    request<AuthResponse>('/auth/login', { method: 'POST', body: JSON.stringify(payload) }),

  logout: () => request<{ status: string }>('/auth/logout', { method: 'POST' }),

  me: () => request<{ user: User }>('/auth/me'),

  updateMe: (payload: { display_name?: string; old_password?: string; new_password?: string }) =>
    request<{ user: User }>('/auth/me', { method: 'PUT', body: JSON.stringify(payload) }),

  myBanks: () => request<{ available: BankInfo[]; selected: string[] }>('/me/banks'),

  saveMyBanks: (selected: string[]) =>
    request<{ selected: string[] }>('/me/banks', { method: 'PUT', body: JSON.stringify({ selected }) }),

  // --- 教练端：用户管理 / 题库动态加载 ---
  adminUsers: () => request<AdminUser[]>('/admin/users'),

  adminCreateUser: (payload: { username: string; password: string; display_name?: string }) =>
    request<User>('/admin/users', { method: 'POST', body: JSON.stringify(payload) }),

  adminUpdateUser: (id: string, payload: { display_name?: string; disabled?: boolean; password?: string }) =>
    request<User>(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),

  adminDeleteUser: (id: string) =>
    request<{ status: string }>(`/admin/users/${id}`, { method: 'DELETE' }),

  adminBanks: () => request<BankSource[]>('/admin/banks'),

  adminUploadBank: (payload: { name: string; content: string }) =>
    request<{ name: string; imported: number }>('/admin/banks', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  createExercise: (ex: ExerciseDetail) =>
    request<ExerciseDetail>('/admin/exercises', { method: 'POST', body: JSON.stringify(ex) }),

  updateExercise: (id: string, ex: Partial<ExerciseDetail>) =>
    request<ExerciseDetail>(`/admin/exercises/${id}`, { method: 'PUT', body: JSON.stringify(ex) }),

  deleteExercise: (id: string) =>
    request<{ status: string }>(`/admin/exercises/${id}`, { method: 'DELETE' }),

  exportExercises: () => request<ExerciseDetail[]>('/admin/export'),

  importExercises: (list: ExerciseDetail[]) =>
    request<{ imported: number; total: number }>('/admin/import', {
      method: 'POST',
      body: JSON.stringify(list),
    }),

  categories: () => request<CategoryItem[]>('/admin/categories'),

  createCategory: (tech: string, sub: string) =>
    request<unknown>('/admin/categories', {
      method: 'POST',
      body: JSON.stringify({ tech_stack: tech, subcategory: sub }),
    }),

  renameCategory: (tech: string, sub: string, newSub: string) =>
    request<unknown>('/admin/categories', {
      method: 'PUT',
      body: JSON.stringify({ tech_stack: tech, subcategory: sub, new_subcategory: newSub }),
    }),

  deleteCategory: (tech: string, sub: string) =>
    request<unknown>('/admin/categories', {
      method: 'DELETE',
      body: JSON.stringify({ tech_stack: tech, subcategory: sub }),
    }),

  aiGenerate: (payload: { tech_stack: string; weight: string; knowledge_points: string; count: number }) =>
    request<AiGenResponse>('/admin/ai-generate', { method: 'POST', body: JSON.stringify(payload) }),

  llmSettings: () => request<LLMSettings>('/admin/llm-settings'),

  saveLLMSettings: (payload: { endpoint: string; model: string }) =>
    request<{ ok: boolean }>('/admin/llm-settings', { method: 'PUT', body: JSON.stringify(payload) }),

  testLLMSettings: (payload: { endpoint: string; model: string }) =>
    request<{ ok: boolean; models: string[]; endpoint: string }>('/admin/llm-settings/test', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  adminEnvironments: () => request<Environment[]>('/admin/environments'),

  createEnvironment: (env: Partial<Environment>) =>
    request<Environment>('/admin/environments', { method: 'POST', body: JSON.stringify(env) }),

  deleteEnvironment: (id: string) =>
    request<{ status: string }>(`/admin/environments/${id}`, { method: 'DELETE' }),

  testEnvironment: (id: string) =>
    request<{ id: string; health_status: string; detail: string }>(`/admin/environments/${id}/test`, {
      method: 'POST',
    }),

  adminStats: () => request<Record<string, unknown>>('/admin/stats'),
};
