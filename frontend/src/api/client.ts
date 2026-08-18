import type {
  CategoryItem,
  Environment,
  ExerciseDetail,
  ExerciseListItem,
  ExecuteResult,
  FavoriteItem,
  ProgressResponse,
  Variant,
  WeakPointItem,
} from '../types';
const BASE = '/api/v1';

export interface AiGenResponse {
  total: number;
  ok: number;
  items: { id: string; title: string; ok: boolean; error?: string }[];
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
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
