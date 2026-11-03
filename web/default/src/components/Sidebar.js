import React, { useContext, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Dropdown, Icon } from 'semantic-ui-react';
import { UserContext } from '../context/User';
import { API, getLogo, getSystemName, isAdmin, showSuccess } from '../helpers';
import './Sidebar.css';

// Console navigation. Every entry is a route inside the console; the
// landing page lives outside of it and has its own header.
const NAV_ITEMS = [
  { name: 'header.dashboard', to: '/dashboard', icon: 'chart bar' },
  { name: 'header.token', to: '/token', icon: 'key' },
  { name: 'header.topup', to: '/topup', icon: 'wallet' },
  { name: 'header.log', to: '/log', icon: 'history' },
  { name: 'header.setting', to: '/setting', icon: 'setting' },
];

const ADMIN_ITEMS = [
  { name: 'header.channel', to: '/channel', icon: 'sitemap' },
  { name: 'header.redemption', to: '/redemption', icon: 'gift' },
  { name: 'header.user', to: '/user', icon: 'users' },
];

const LANGUAGE_OPTIONS = [
  { key: 'en', text: 'English', value: 'en' },
  { key: 'id', text: 'Bahasa Indonesia', value: 'id' },
];

const Sidebar = () => {
  const { t, i18n } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(false);

  const systemName = getSystemName();
  const logo = getLogo();

  // i18next can report a regional tag such as "en-US"; the dropdown only
  // knows about "en" and "id", so normalise before matching.
  const currentLanguage = LANGUAGE_OPTIONS.some(
    (o) => o.value === i18n.language
  )
    ? i18n.language
    : i18n.language.slice(0, 2);

  const isActive = (to) => location.pathname === to;

  const go = (to) => {
    navigate(to);
    setOpen(false);
  };

  const logout = async () => {
    setOpen(false);
    await API.get('/api/user/logout');
    showSuccess('Signed out!');
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  };

  const renderItem = (item) => {
    const active = isActive(item.to);
    return (
      <button
        key={item.to}
        type='button'
        className={`sidebar-item ${active ? 'is-active' : ''}`}
        onClick={() => go(item.to)}
      >
        <Icon name={item.icon} />
        <span>{t(item.name)}</span>
      </button>
    );
  };

  return (
    <>
      <aside className={`sidebar ${open ? 'is-open' : ''}`}>
        <div className='sidebar-brand'>
          <Link to='/' className='sidebar-brand-link' onClick={() => setOpen(false)}>
            <img src={logo} alt='logo' />
            <span className='sidebar-brand-name'>{systemName}</span>
          </Link>
          <button
            type='button'
            className='sidebar-close'
            onClick={() => setOpen(false)}
            aria-label='Close'
          >
            <Icon name='close' />
          </button>
        </div>

        <nav className='sidebar-nav'>
          <div className='sidebar-group'>{NAV_ITEMS.map(renderItem)}</div>

          {isAdmin() && (
            <div className='sidebar-group'>
              <div className='sidebar-group-label'>{t('sidebar.group_admin')}</div>
              {ADMIN_ITEMS.map(renderItem)}
            </div>
          )}

          <div className='sidebar-group'>
            <div className='sidebar-group-label'>{t('sidebar.group_site')}</div>
            <button
              type='button'
              className='sidebar-item'
              onClick={() => go('/about')}
            >
              <Icon name='info circle' />
              <span>{t('header.about')}</span>
            </button>
          </div>
        </nav>

        <div className='sidebar-footer'>
          <Dropdown
            selection
            fluid
            className='sidebar-language'
            options={LANGUAGE_OPTIONS}
            value={currentLanguage}
            onChange={(_, { value }) => i18n.changeLanguage(value)}
          />

          {userState.user ? (
            <div className='sidebar-user'>
              <div className='sidebar-user-name'>{userState.user.username}</div>
              <button type='button' className='sidebar-logout' onClick={logout}>
                <Icon name='sign out' />
                <span>{t('header.logout')}</span>
              </button>
            </div>
          ) : (
            <button
              type='button'
              className='sidebar-item'
              onClick={() => go('/login')}
            >
              <Icon name='sign in' />
              <span>{t('header.login')}</span>
            </button>
          )}
        </div>
      </aside>

      {open && <div className='sidebar-overlay' onClick={() => setOpen(false)} />}

      <button
        type='button'
        className='sidebar-toggle'
        onClick={() => setOpen(!open)}
        aria-label='Toggle navigation'
      >
        <Icon name={open ? 'close' : 'bars'} />
      </button>
    </>
  );
};

export default Sidebar;
