// Theme handling for the whole app.
//
// The mode lives in localStorage under THEME_STORAGE_KEY and is applied as a
// class on <html>, so every stylesheet can react to it with
// `html.dark` / `html.light` selectors.
//
// `system` follows the OS preference and keeps following it until the user
// picks light or dark explicitly.

export const THEME_STORAGE_KEY = 'panway_theme';

export const THEME_MODES = ['light', 'dark', 'system'];

const listeners = new Set();

export function resolveTheme(mode) {
  if (mode === 'light' || mode === 'dark') {
    return mode;
  }
  if (
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-color-scheme: dark)').matches
  ) {
    return 'dark';
  }
  return 'light';
}

export function getThemeMode() {
  if (typeof localStorage === 'undefined') {
    return 'system';
  }
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  return THEME_MODES.includes(stored) ? stored : 'system';
}

export function getResolvedTheme() {
  return resolveTheme(getThemeMode());
}

export function applyTheme(mode) {
  const resolved = resolveTheme(mode);
  if (typeof document === 'undefined') {
    return resolved;
  }
  const root = document.documentElement;
  root.classList.remove('light', 'dark');
  root.classList.add(resolved);
  root.style.colorScheme = resolved;
  return resolved;
}

export function setThemeMode(mode) {
  const next = THEME_MODES.includes(mode) ? mode : 'system';
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(THEME_STORAGE_KEY, next);
  }
  applyTheme(next);
  listeners.forEach((fn) => fn(next));
  return next;
}

export function toggleTheme() {
  const next = getResolvedTheme() === 'dark' ? 'light' : 'dark';
  return setThemeMode(next);
}

export function onThemeChange(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

// Keep the "system" mode in sync with the OS while the app is open.
if (typeof window !== 'undefined' && window.matchMedia) {
  const mq = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = () => {
    if (getThemeMode() === 'system') {
      applyTheme('system');
    }
  };
  if (mq.addEventListener) {
    mq.addEventListener('change', handler);
  }
}
