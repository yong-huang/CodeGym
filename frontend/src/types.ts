export type ExerciseType = 'fill' | 'bugfix' | 'command' | 'config';
export type WeightLevel = 'beginner' | 'easy' | 'medium' | 'hard' | 'expert';

export interface ExerciseListItem {
  id: string;
  title: string;
  description: string;
  exercise_type: ExerciseType;
  tech_stack: string;
  weight: string;
  category: string;
  subcategory: string;
  point_count: number;
}

export interface VariantFillPoint {
  id: string;
  category: string;
  fill_type: string;
  line: number;
  options?: string[];
  hint_level_1: string;
  hint_level_2: string;
  hint_level_3: string;
  hint_level_4: string;
}

export interface SpotterHints {
  level_1: string;
  level_2: string;
  level_3: string;
  level_4: string;
}

export interface Variant {
  exercise_id: string;
  title: string;
  description: string;
  exercise_type: ExerciseType;
  tech_stack: string;
  category: string;
  weight_level: WeightLevel;
  variant_id: string;
  marked_code: string;
  code_for_edit?: string;
  bug_description?: string;
  fill_points: VariantFillPoint[];
  spotter_hints: SpotterHints;
  time_limit_seconds: number;
}

export interface FillResult {
  id: string;
  point_id: string;
  user_answer: string;
  expected: string;
  is_correct: boolean;
  score: number;
}

export interface ExecuteResult {
  execution_id: string;
  success: boolean;
  output: string;
  expected_output: string;
  is_correct: boolean;
  score: number;
  grade: string;
  fill_results: FillResult[];
  execution_time_ms: number;
  logs: string[];
  weak_analysis?: string;
  error_message?: string;
  correct_answer?: string;
  explanation?: string;
}

export interface Progress {
  user_id: string;
  total_attempts: number;
  correct_attempts: number;
  completed_exercises: Record<string, boolean>;
  weak_points: Record<string, number>;
  tech_stack_scores: Record<string, number>;
  total_hint_used: number;
  current_streak: number;
  longest_streak: number;
}

export interface WeakPointItem {
  exercise_id: string;
  title: string;
  fail_count: number;
}

export interface FavoriteItem {
  exercise_id: string;
  title: string;
  tech_stack: string;
  exercise_type: string;
  favorited_at: string;
}

export interface CategoryItem {
  tech_stack: string;
  subcategory: string;
  exercise_count: number;
}

export interface Environment {
  id: string;
  name: string;
  type: string;
  config: Record<string, unknown>;
  health_status: string;
  max_concurrent: number;
  enabled: boolean;
  created_at: string;
}

export interface RecentAttempt {
  id: string;
  exercise_id: string;
  weight_level: string;
  score: number;
  grade: string;
  is_correct: boolean;
  created_at: string;
}

export interface ProgressResponse {
  progress: Progress;
  recent_attempts: RecentAttempt[];
}

export interface ExerciseDetail {
  id: string;
  title: string;
  description: string;
  exercise_type: ExerciseType;
  tech_stack: string;
  weight: string;
  category: string;
  subcategory: string;
  full_code: string;
  buggy_code?: string;
  bug_description?: string;
  bug_location?: string;
  fill_points: {
    points: { id: string; original_text: string; location: { span_start: number; span_end: number } }[];
  };
  expected_output?: string;
  test_cases?: { input?: string; output: string }[];
  spotter_hints: SpotterHints;
  time_limit_seconds: number;
  explanation?: string;
  enabled: boolean;
}
