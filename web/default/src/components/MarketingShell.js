import React, { useContext, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../context/User';
import { API, getLogo, getSystemName } from '../helpers';
import { renderQuota } from '../helpers/render';
import ThemeToggle from './ThemeToggle';
import './MarketingShell.css';

// Top navigation for the public pages (landing, models, pricing, about).
// It is deliberately not the console sidebar: a visitor who has not signed
// in yet has nothing to navigate to inside the console.

const MarketingHeader = ({ current }) => {
  const { t } = useTranslation();
  const [userState] = useContext(UserContext);
  const [quota, setQuota] = useState(null);

  const systemName = getSystemName();
  const logo = getLogo();

  useEffect(() => {
    if (!userState.user) {
      setQuota(null);
      return;
    }
    API.get('/api/user/self')
      .then((res) => {
        if (res.data.success) {
          setQuota(res.data.data.quota);
        }
      })
      .catch(() => {});
  }, [userState.user]);

  // These are public marketing pages. Even when the visitor is signed in,
  // Models/Pricing/About must stay on the marketing route; only the Console
  // button enters the dashboard.
  const links = [
    { to: '/', label: t('landing.nav.home'), key: 'home' },
    { to: '/models', label: t('landing.nav.models'), key: 'models' },
    { to: '/pricing', label: t('landing.nav.pricing'), key: 'pricing' },
  ];

  return (
    <header className='marketing-header'>
      <div className='marketing-header-inner'>
        <Link to='/' className='marketing-brand'>
          <img src={logo} alt={systemName} />
          <span>{systemName}</span>
        </Link>

        <nav className='marketing-nav'>
          {links.map((l) => (
            <Link
              key={l.to}
              to={l.to}
              className={current === l.key ? 'is-active' : ''}
            >
              {l.label}
            </Link>
          ))}
        </nav>

        <div className='marketing-actions'>
          {userState.user && quota !== null && (
            <Link to='/topup' className='marketing-balance'>
              <span className='marketing-balance-label'>
                {t('landing.header.balance')}
              </span>
              <strong>{renderQuota(quota, t)}</strong>
            </Link>
          )}
          <ThemeToggle />
          <Link
            to={userState.user ? '/dashboard' : '/login'}
            className='marketing-btn'
          >
            {userState.user ? t('landing.nav.console') : t('landing.nav.signin')}
          </Link>
        </div>
      </div>
    </header>
  );
};

const MarketingShell = ({ current, children }) => {
  const { t } = useTranslation();
  const [userState] = useContext(UserContext);
  const systemName = getSystemName();

  return (
    <div className='marketing-shell'>
      <MarketingHeader current={current} />

      <main className='marketing-main'>{children}</main>

      <footer className='marketing-footer'>
        <div className='marketing-footer-inner'>
          <span>
            &copy; {new Date().getFullYear()} {systemName}.{' '}
            {t('landing.footer.rights')}
          </span>
<div className='marketing-footer-links'>
        <Link to='/login'>{t('landing.footer.signin')}</Link>
            {userState.user && (
              <Link to='/dashboard'>{t('landing.nav.console')}</Link>
            )}
          </div>
        </div>
      </footer>
    </div>
  );
};

export default MarketingShell;
