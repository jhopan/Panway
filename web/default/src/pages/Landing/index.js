import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import { getLogo, getSystemName } from '../../helpers';
import MarketingShell from '../../components/MarketingShell';
import './Landing.css';

// Providers Panway can front. The orbit visual is the centrepiece of the
// hero, so the list lives here.
const PROVIDERS = [
  { key: 'openai', name: 'OpenAI' },
  { key: 'claude', name: 'Claude' },
  { key: 'gemini', name: 'Gemini' },
  { key: 'deepseek', name: 'DeepSeek' },
  { key: 'qwen', name: 'Qwen' },
  { key: 'minimax', name: 'MiniMax' },
  { key: 'kimi', name: 'Kimi' },
  { key: 'grok', name: 'Grok' },
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
  const formatPrice = (value) =>
    'Rp' + Math.round(value).toLocaleString(locale);

  const steps = STEPS.map((s) => ({
    no: s.no,
    title: t(`${s.key}.title`),
    body: t(`${s.key}.body`),
  }));

  // Signed in customers get the console links; visitors get sign in.
  // The console link lives in the action button on the right, so it is not
  // repeated in the nav list.
  const navLinks = [
    { to: '/models', label: t('landing.nav.models') },
    { to: '/pricing', label: t('landing.nav.pricing') },
    { to: '/about', label: t('landing.nav.about') },
  ];

  return (
    <MarketingShell current='home'>
      <div className='landing'>
      <section className='landing-hero'>
        <div className='landing-hero-inner'>
          <div className='landing-hero-copy'>
            <div className='landing-eyebrow'>
              <span className='landing-eyebrow-rule' />
              {t('landing.hero.eyebrow')}
            </div>

            <h1 className='landing-hero-title'>
              <span className='landing-hero-title-accent'>
                {t('landing.hero.title_accent')}
              </span>
              <br />
              {t('landing.hero.title')}
            </h1>

            <p className='landing-hero-sub'>{t('landing.hero.subtitle')}</p>

            <div className='landing-hero-actions'>
              <button
                className='landing-btn landing-btn-primary landing-btn-lg'
                onClick={() => navigate(consolePath)}
              >
                {t('landing.hero.cta_primary')}
              </button>
              <Link
                className='landing-btn landing-btn-ghost landing-btn-lg'
                to='/models'
              >
                {t('landing.hero.cta_secondary')}
              </Link>
            </div>

            <ul className='landing-checks'>
              <li>{t('landing.hero.check1')}</li>
              <li>{t('landing.hero.check2')}</li>
              <li>{t('landing.hero.check3')}</li>
            </ul>
          </div>

          <div className='landing-hero-visual' aria-hidden='true'>
            <div className='landing-hub'>
              <div className='landing-hub-ring' />
              <div className='landing-hub-ring landing-hub-ring-outer' />
              <div className='landing-hub-core'>
                <span className='landing-hub-logo'>{systemName.charAt(0)}</span>
              </div>
              <div className='landing-hub-orbit'>
                {PROVIDERS.map((p, i) => (
                  <span
                    key={p.key}
                    className='landing-hub-node'
                    style={{ '--i': i, '--n': PROVIDERS.length }}
                  >
                    {p.name}
                  </span>
                ))}
              </div>
              <div className='landing-hub-caption'>
                {t('landing.hero.hub_caption')}
              </div>
            </div>
          </div>
        </div>

        <div className='landing-hero-strip'>
          <div className='landing-section-inner landing-hero-strip-inner'>
            <span>{t('landing.hero.strip_left')}</span>
            <span>{t('landing.hero.strip_right')}</span>
          </div>
        </div>
      </section>

      <section className='landing-logos'>
        <div className='landing-section-inner'>
          <div className='landing-eyebrow landing-eyebrow-center'>
            <span className='landing-eyebrow-rule' />
            {t('landing.logos.eyebrow')}
            <span className='landing-eyebrow-rule' />
          </div>
          <div className='landing-logos-row'>
            {PROVIDERS.map((p) => (
              <div key={p.key} className='landing-logo-item'>
                <span className='landing-logo-name'>{p.name}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className='landing-models' id='models'>
        <div className='landing-section-inner'>
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
              <Link key={m.model_id} to='/models' className='landing-model-card'>
                <div className='landing-model-card-head'>
                  {m.provider && (
                    <span className='landing-model-provider'>
                      {m.provider}
                    </span>
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
              </Link>
            ))}
          </div>

          <div className='landing-section-more'>
            <Link to='/models' className='landing-btn landing-btn-ghost'>
              {t('landing.models.all')}
            </Link>
          </div>
        </div>
      </section>

      <section className='landing-steps-section' id='how'>
        <div className='landing-section-inner'>
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
        </div>
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

      </div>
    </MarketingShell>
  );
};

export default Landing;
