import { useEffect, useState } from 'react';
import HomePage from './pages/HomePage';
import TrainingPage from './pages/TrainingPage';
import WeakPointsPage from './pages/WeakPointsPage';
import AdminPage from './pages/AdminPage';
import LibraryPage from './pages/LibraryPage';
import LoginPage from './pages/LoginPage';
import BanksPage from './pages/BanksPage';
import Sidebar from './components/Sidebar';
import { logout, refreshUser, useAuthReady, useUser } from './auth';

type Route =
  | { name: 'home' }
  | { name: 'training'; id: string }
  | { name: 'weak-points' }
  | { name: 'library' }
  | { name: 'admin' }
  | { name: 'banks' }
  | { name: 'login' };

const COLLAPSE_KEY = 'codegym:sidebar-collapsed';

// Routes reachable without a trainee account (coach-side tools + login).
const PUBLIC_ROUTES = new Set<Route['name']>(['admin', 'library', 'login']);

function parseHash(): Route {
  const [path] = window.location.hash.replace(/^#/, '').split('?');
  const parts = path.split('/').filter(Boolean);
  if (parts[0] === 'exercise' && parts[1]) {
    return { name: 'training', id: parts[1] };
  }
  if (parts[0] === 'weak-points') return { name: 'weak-points' };
  if (parts[0] === 'library') return { name: 'library' };
  if (parts[0] === 'admin') return { name: 'admin' };
  if (parts[0] === 'banks') return { name: 'banks' };
  if (parts[0] === 'login') return { name: 'login' };
  return { name: 'home' };
}

function navigate(route: Route) {
  if (route.name === 'home') window.location.hash = '#/';
  else if (route.name === 'training') window.location.hash = `#/exercise/${route.id}`;
  else if (route.name === 'weak-points') window.location.hash = '#/weak-points';
  else if (route.name === 'library') window.location.hash = '#/library';
  else if (route.name === 'admin') window.location.hash = '#/admin';
  else if (route.name === 'banks') window.location.hash = '#/banks';
  else if (route.name === 'login') window.location.hash = '#/login';
}

export function useHashRoute(): Route {
  const [route, setRoute] = useState<Route>(parseHash);
  useEffect(() => {
    const onChange = () => setRoute(parseHash());
    window.addEventListener('hashchange', onChange);
    return () => window.removeEventListener('hashchange', onChange);
  }, []);
  return route;
}

export { navigate };

export default function App() {
  const route = useHashRoute();
  const user = useUser();
  const ready = useAuthReady();
  // Default to collapsed; remember the user's last choice (when set).
  const [collapsed, setCollapsed] = useState<boolean>(() => {
    const stored = localStorage.getItem(COLLAPSE_KEY);
    return stored === null ? true : stored === '1';
  });

  useEffect(() => {
    void refreshUser();
  }, []);

  // Redirect anonymous visitors to the login page for guarded routes.
  useEffect(() => {
    if (ready && !user && !PUBLIC_ROUTES.has(route.name)) {
      navigate({ name: 'login' });
    }
  }, [ready, user, route.name]);

  const toggleSidebar = () => {
    setCollapsed((c) => {
      const next = !c;
      localStorage.setItem(COLLAPSE_KEY, next ? '1' : '0');
      return next;
    });
  };

  const doLogout = async () => {
    await logout();
    navigate({ name: 'login' });
  };

  const showApp = route.name !== 'login' && (user !== null || PUBLIC_ROUTES.has(route.name));

  return (
    <div className="app">
      <header className="topbar">
        <div className="topbar-left">
          <button
            className={`sidebar-hamburger ${collapsed ? 'collapsed' : ''}`}
            onClick={toggleSidebar}
            title={collapsed ? '展开题库目录' : '收起题库目录'}
          >
            ☰
          </button>
          <button className="brand" onClick={() => navigate({ name: 'home' })}>
            🏋️ <strong>CodeGym</strong>
            <span className="slogan">练出来的，才是自己的</span>
          </button>
        </div>
        <nav className="nav">
          <button className="nav-link" onClick={() => navigate({ name: 'home' })}>
            📚 训练计划
          </button>
          <button className="nav-link" onClick={() => navigate({ name: 'banks' })}>
            🧪 我的题库
          </button>
          <button className="nav-link" onClick={() => navigate({ name: 'weak-points' })}>
            📝 薄弱部位
          </button>
          <button className="nav-link" onClick={() => navigate({ name: 'library' })}>
            📦 题库管理
          </button>
          <button className="nav-link" onClick={() => navigate({ name: 'admin' })}>
            🧑‍🏫 教练端
          </button>
          {user ? (
            <span className="user-chip">
              👤 <span className="user-name">{user.display_name || user.username}</span>
              <button className="nav-link user-logout" onClick={() => void doLogout()}>
                退出
              </button>
            </span>
          ) : (
            ready && (
              <button className="nav-link" onClick={() => navigate({ name: 'login' })}>
                🔑 登录
              </button>
            )
          )}
        </nav>
      </header>

      <div className="body">
        {showApp && (
          <Sidebar
            collapsed={collapsed}
            activeExerciseId={route.name === 'training' ? route.id : undefined}
          />
        )}
        <main className="main">
          {route.name === 'login' && <LoginPage />}
          {route.name === 'home' && user && <HomePage />}
          {route.name === 'training' && user && <TrainingPage id={route.id} />}
          {route.name === 'weak-points' && user && <WeakPointsPage />}
          {route.name === 'banks' && user && <BanksPage />}
          {route.name === 'library' && <LibraryPage />}
          {route.name === 'admin' && <AdminPage />}
        </main>
      </div>

      <footer className="footer">💪 练出来的，才是自己的 · CodeGym</footer>
    </div>
  );
}
