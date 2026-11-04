import React, { useEffect, useState } from 'react';
import { Icon } from 'semantic-ui-react';
import {
  getResolvedTheme,
  getThemeMode,
  onThemeChange,
  setThemeMode,
  THEME_MODES,
} from '../helpers/theme';
import './ThemeToggle.css';

// Sun / moon switch. Clicking cycles light -> dark -> system.
const NEXT_MODE = {
  light: 'dark',
  dark: 'system',
  system: 'light',
};

const LABEL = {
  light: 'Light theme',
  dark: 'Dark theme',
  system: 'System theme',
};

const ICON = {
  light: 'sun',
  dark: 'moon',
  system: 'circle half',
};

const ThemeToggle = () => {
  const [mode, setMode] = useState(getThemeMode());

  useEffect(() => {
    return onThemeChange((next) => setMode(next));
  }, []);

  const cycle = () => {
    const next = NEXT_MODE[getThemeMode()] || 'light';
    setThemeMode(next);
    setMode(next);
  };

  const resolved = getResolvedTheme();

  return (
    <button
      type='button'
      className='theme-toggle'
      onClick={cycle}
      title={`${LABEL[mode]}${
        mode === 'system' ? ` (${resolved})` : ''
      } - click to switch`}
      aria-label={LABEL[mode]}
      data-mode={mode}
    >
      <Icon name={ICON[mode]} />
    </button>
  );
};

export default ThemeToggle;
