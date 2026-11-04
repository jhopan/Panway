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
// Dark is the default, so the cycle starts there and ends back on it.
const NEXT_MODE = {
  dark: 'light',
  light: 'system',
  system: 'dark',
};

const LABEL = {
  light: 'Light theme',
  dark: 'Dark theme',
  system: 'System theme',
};

const ICON = {
  light: 'sun',
  dark: 'moon',
  // "circle half" does not exist in the Semantic icon font and rendered as
  // an empty blue box. "adjust" does exist and reads as "auto".
  system: 'adjust',
};

const ThemeToggle = () => {
  const [mode, setMode] = useState(getThemeMode());

  useEffect(() => {
    return onThemeChange((next) => setMode(next));
  }, []);

  const cycle = () => {
    const next = NEXT_MODE[getThemeMode()] || 'dark';
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
