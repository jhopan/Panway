import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import { getLogo, getSystemName } from '../../helpers';
import './Landing.css';

// A small console preview: what a customer actually cares about is seeing
// their usage, so the hero shows the product instead of a logo wall.
const PREVIEW_ROWS = [
  { model: 'agnes/agnes-3.0-flash', tokens: '12.4k', ms: '842' },
  { model: 'th/deepseek-v4-flash:free', tokens: '3.1k', ms: '390' },
  {
    model: 'cf/@cf/qwen/qwen2.5-coder-32b-instruct',
    tokens: '8.9k',
    ms: '1.2s',
  },
];

const STEPS = [
  { no: '01', key: 'landing.how.steps.connect' },
  { no: '02', key: 'landing.how.steps.topup' },
  { no: '03', key: 'landing.how.steps.call' },
];

const Landing = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [userState] = useContext(UserContext);
  const [models, setModels] = useState([]);
  const [pricingError, setPricingError] = useState(false);

  const logo = getLogo();
  const systemName = getSystemName();
  const locale = i18n.language === 'id' ? 'id-ID' : 'en-US';

  useEffect(() => {
    document.title = systemName;
    fetch('/api/pricing')
      .then((res) => res.json())
      .then((body) => {
        if (body.success && body.data && Array.isArray(body.data.models)) {
          setModels(body.data.models);
        } else {
          setPricingError(true);
        }
      })
      .catch(() => setPricingError(true));
  }, [systemName]);

  const consolePath = userState.user ? '/dashboard' : '/login';

  const formatPrice = (value) => 'Rp' + Math.round(value).toLocaleString(locale);

  const steps = STEPS.map((s) => ({
    no: s.no,
    title: t(`${s.key}.title`),
    body: t(`${s.key}.body`),
  }));

  return (
    <div className='landing'>
      <header className='landing-header'>
        <div className='landing-header-inner'>
          <Link to='/' className='landing-brand'>
            <img src={logo} alt={systemName} />
            <span>{systemName}</span>
          </Link>

          <nav className='landing-nav'>
            <a href='#models'>{t('landing.nav.models')}</a>
            <a href='#how'>{t('landing.nav.how')}</a>
            <a href='#pricing'>{t('landing.nav.pricing')}</a>
          </nav>

          <div className='landing-header-actions'>
            <Link to={consolePath} className='landing-btn landing-btn-primary'>
              {userState.user ? t('landing.nav.console') : t('landing.nav.signin')}
            </Link>
          </div>
        </div>
      </header>

      <section className='landing-hero'>
        <div className='landing-hero-inner'>
          <div className='landing-hero-copy'>
            <span className='landing-pill'>{t('landing.hero.pill')}</span>

            <h1 className='landing-hero-title'>{t('landing.hero.title')}</h1>

            <p className='landing-hero-sub'>{t('landing.hero.subtitle')}</p>

            <div className='landing-hero-actions'>
              <button
                className='landing-btn landing-btn-primary landing-btn-lg'
                onClick={() => navigate(consolePath)}
              >
                {t('landing.hero.cta_primary')}
              </button>
              <a
                className='landing-btn landing-btn-ghost landing-btn-lg'
                href='#models'
              >
                {t('landing.hero.cta_secondary')}
              </a>
            </div>

            <div className='landing-hero-stats'>
              <div className='landing-hero-stat'>
                <span className='landing-hero-stat-value'>
                  {models.length > 0 ? models.length : '-'}
                </span>
                <span className='landing-hero-stat-label'>
                  {t('landing.hero.stat_models')}
                </span>
              </div>
              <div className='landing-hero-stat'>
                <span className='landing-hero-stat-value'>1</span>
                <span className='landing-hero-stat-label'>
                  {t('landing.hero.stat_key')}
                </span>
              </div>
              <div className='landing-hero-stat'>
                <span className='landing-hero-stat-value'>IDR</span>
                <span className='landing-hero-stat-label'>
                  {t('landing.hero.stat_currency')}
                </span>
              </div>
            </div>
          </div>

          <div className='landing-hero-panel' aria-hidden='true'>
            <div className='landing-panel'>
              <div className='landing-panel-bar'>
                <span className='landing-panel-title'>
                  {t('landing.preview.title')}
                </span>
                <span className='landing-panel-dot' />
                <span className='landing-panel-dot' />
                <span className='landing-panel-dot' />
              </div>

              <div className='landing-panel-body'>
                <div className='landing-panel-chart'>
                  {[38, 62, 45, 78, 55, 88, 70, 96, 82, 100].map((h, i) => (
                    <span
                      key={i}
                      className='landing-panel-bar-fill'
                      style={{ height: `${h}%` }}
                    />
                  ))}
                </div>

                <div className='landing-panel-rows'>
                  {PREVIEW_ROWS.map((row) => (
                    <div key={row.model} className='landing-panel-row'>
                      <span className='landing-panel-model'>{row.model}</span>
                      <span className='landing-panel-tokens'>{row.tokens}</span>
                      <span className='landing-panel-ms'>{row.ms}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className='landing-section' id='models'>
        <div className='landing-section-head'>
          <h2 className='landing-section-title'>{t('landing.models.title')}</h2>
          <p className='landing-section-lead'>{t('landing.models.lead')}</p>
        </div>

        <div className='landing-model-grid'>
          {models.length === 0 && !pricingError && (
            <div className='landing-empty'>{t('landing.models.loading')}</div>
          )}
          {models.length === 0 && pricingError && (
            <div className='landing-empty'>{t('landing.models.error')}</div>
          )}
          {models.map((m) => (
            <div key={m.model_id} className='landing-model-card'>
              <div className='landing-model-card-head'>
                {m.provider && (
                  <span className='landing-model-provider'>{m.provider}</span>
                )}
                <span className='landing-model-id'>{m.name}</span>
              </div>

              <div className='landing-model-prices'>
                <div className='landing-model-price'>
                  <span className='landing-model-price-label'>
                    {t('landing.models.input')}
                  </span>
                  <span className='landing-model-price-value'>
                    {formatPrice(m.input_price)}
                  </span>
                </div>
                <div className='landing-model-price'>
                  <span className='landing-model-price-label'>
                    {t('landing.models.output')}
                  </span>
                  <span className='landing-model-price-value'>
                    {formatPrice(m.output_price)}
                  </span>
                </div>
              </div>

              <div className='landing-model-unit'>{t('landing.models.unit')}</div>

              <code className='landing-model-code'>{m.model_id}</code>
            </div>
          ))}
        </div>
      </section>

      <section className='landing-section landing-section-alt' id='how'>
        <div className='landing-section-head'>
          <h2 className='landing-section-title'>{t('landing.how.title')}</h2>
          <p className='landing-section-lead'>{t('landing.how.lead')}</p>
        </div>

        <div className='landing-steps'>
          {steps.map((step, i) => (
            <div key={step.no} className='landing-step'>
              <div className='landing-step-top'>
                <span className='landing-step-no'>{step.no}</span>
                {i < steps.length - 1 && (
                  <span className='landing-step-line' aria-hidden='true' />
                )}
              </div>
              <h3 className='landing-step-title'>{step.title}</h3>
              <p className='landing-step-body'>{step.body}</p>
            </div>
          ))}
        </div>
      </section>

      <section className='landing-section' id='pricing'>
        <div className='landing-section-head'>
          <h2 className='landing-section-title'>{t('landing.pricing.title')}</h2>
          <p className='landing-section-lead'>{t('landing.pricing.lead')}</p>
        </div>

        <div className='landing-table-wrap'>
          <table className='landing-table'>
            <thead>
              <tr>
                <th>{t('landing.models.model')}</th>
                <th>{t('landing.models.input')}</th>
                <th>{t('landing.models.output')}</th>
                <th>{t('landing.pricing.multiplier')}</th>
              </tr>
            </thead>
            <tbody>
              {models.length === 0 && (
                <tr>
                  <td colSpan={4} className='landing-empty'>
                    {t('landing.models.loading')}
                  </td>
                </tr>
              )}
              {models.map((m) => (
                <tr key={m.model_id}>
                  <td>
                    <code>{m.model_id}</code>
                  </td>
                  <td>{formatPrice(m.input_price)}</td>
                  <td>{formatPrice(m.output_price)}</td>
                  <td className='landing-table-dim'>
                    {m.output_price / m.input_price}x
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <p className='landing-note'>{t('landing.pricing.note')}</p>
      </section>

      <section className='landing-cta'>
        <div className='landing-cta-inner'>
          <h2 className='landing-cta-title'>{t('landing.cta.title')}</h2>
          <p className='landing-cta-body'>{t('landing.cta.body')}</p>
          <button
            className='landing-btn landing-btn-primary landing-btn-lg'
            onClick={() => navigate(consolePath)}
          >
            {t('landing.cta.button')}
          </button>
        </div>
      </section>

      <footer className='landing-footer'>
        <div className='landing-footer-inner'>
          <span>
            &copy; {new Date().getFullYear()} {systemName}.{' '}
            {t('landing.footer.rights')}
          </span>
          <div className='landing-footer-links'>
            <Link to='/about'>{t('landing.footer.about')}</Link>
            <Link to='/login'>{t('landing.footer.signin')}</Link>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Landing;
