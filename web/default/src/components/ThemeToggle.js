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
  // "circle half" does not exist in the Semantic icon font and rendered as
  // an empty blue box. "adjust" renders, but reads as a control rather than
  // a theme, so the system mode uses the sun icon and relies on the
  // tooltip to say it follows the OS.
  system: 'sun',
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
