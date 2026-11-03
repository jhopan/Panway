import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import { StatusContext } from '../../context/Status';
import { getLogo, getSystemName } from '../../helpers';
import './Landing.css';

// Providers shown in the logo strip and the model selector.
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

// Models Panway sells, with the panel the code snippet switches on.
const MODELS = [
  { key: 'openai', name: 'OpenAI', model: 'agnes/agnes-3.0-flash', snippet: 'openai' },
  { key: 'claude', name: 'Claude', model: 'th/deepseek-v4-flash:free', snippet: 'anthropic' },
  { key: 'gemini', name: 'Gemini', model: 'cf/@cf/qwen/qwen2.5-coder-32b-instruct', snippet: 'gemini' },
  { key: 'deepseek', name: 'DeepSeek', model: 'deepseek-v4-flash', snippet: 'openai' },
];

const SNIPPETS = {
  openai: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-panway-key",
    base_url="https://your-panway-domain/v1",
)

response = client.chat.completions.create(
    model="agnes/agnes-3.0-flash",
    messages=[{"role": "user", "content": "Hello!"}],
)

print(response.choices[0].message.content)`,
  anthropic: `import anthropic

client = anthropic.Anthropic(
    api_key="sk-your-panway-key",
    base_url="https://your-panway-domain/v1",
)

message = client.messages.create(
    model="th/deepseek-v4-flash:free",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello!"}],
)

print(message.content[0].text)`,
  gemini: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-panway-key",
    base_url="https://your-panway-domain/v1",
)

response = client.chat.completions.create(
    model="cf/@cf/qwen/qwen2.5-coder-32b-instruct",
    messages=[{"role": "user", "content": "Hello!"}],
)

print(response.choices[0].message.content)`,
};

const Landing = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [userState] = useContext(UserContext);
  const [statusState] = useContext(StatusContext);
  const [activeModel, setActiveModel] = useState(MODELS[0]);

  const logo = getLogo();
  const systemName = getSystemName();

  useEffect(() => {
    document.title = systemName;
  }, [systemName]);

  const consolePath = userState.user ? '/dashboard' : '/login';

  const features = [
    {
      no: '01',
      tag: t('landing.features.routing.tag'),
      title: t('landing.features.routing.title'),
      body: t('landing.features.routing.body'),
    },
    {
      no: '02',
      tag: t('landing.features.visibility.tag'),
      title: t('landing.features.visibility.title'),
      body: t('landing.features.visibility.body'),
    },
    {
      no: '03',
      tag: t('landing.features.payments.tag'),
      title: t('landing.features.payments.title'),
      body: t('landing.features.payments.body'),
    },
  ];

  return (
    <div className='landing'>
      <header className='landing-header'>
        <div className='landing-container landing-header-inner'>
          <Link to='/' className='landing-brand'>
            {logo ? <img src={logo} alt={systemName} /> : <span className='landing-brand-mark' />}
            <span>{systemName}</span>
          </Link>

          <nav className='landing-nav'>
            <Link to='/'>{t('landing.nav.home')}</Link>
            {!userState.user && (
              <Link to='/login'>{t('landing.nav.console')}</Link>
            )}
            <Link to='/topup'>{t('landing.nav.pricing')}</Link>
            <Link to='/about'>{t('landing.nav.about')}</Link>
          </nav>

          <div className='landing-header-actions'>
            <Link to={consolePath} className='landing-btn landing-btn-primary'>
              {userState.user ? t('landing.nav.console') : t('landing.nav.signin')}
            </Link>
          </div>
        </div>
      </header>

      <section className='landing-hero'>
        <div className='landing-container landing-hero-inner'>
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
              <a className='landing-btn landing-btn-ghost landing-btn-lg' href='#models'>
                {t('landing.hero.cta_secondary')}
              </a>
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
          <div className='landing-container landing-hero-strip-inner'>
            <span>{t('landing.hero.strip_left')}</span>
            <span>{t('landing.hero.strip_right')}</span>
          </div>
        </div>
      </section>

      <section className='landing-logos'>
        <div className='landing-container'>
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

      <section className='landing-code' id='models'>
        <div className='landing-container landing-code-card'>
          <div className='landing-code-copy'>
            <div className='landing-eyebrow'>
              <span className='landing-eyebrow-rule' />
              {t('landing.code.eyebrow')}
            </div>
            <h2 className='landing-section-title'>{t('landing.code.title')}</h2>
            <p className='landing-section-body'>{t('landing.code.body')}</p>

            <div className='landing-model-tabs'>
              {MODELS.map((m) => (
                <button
                  key={m.key}
                  className={`landing-model-tab ${
                    activeModel.key === m.key ? 'is-active' : ''
                  }`}
                  onClick={() => setActiveModel(m)}
                  title={m.model}
                >
                  {m.name}
                </button>
              ))}
            </div>

            <div className='landing-model-id'>
              <code>{activeModel.model}</code>
            </div>
          </div>

          <div className='landing-terminal'>
            <div className='landing-terminal-bar'>
              <span className='landing-terminal-dots'>
                <i />
                <i />
                <i />
              </span>
              <span className='landing-terminal-file'>connection.py</span>
              <span className='landing-terminal-lang'>Python</span>
            </div>
            <pre className='landing-terminal-body'>
              <code>{SNIPPETS[activeModel.snippet]}</code>
            </pre>
            <div className='landing-terminal-foot'>
              <span>PROVIDER / {activeModel.name.toUpperCase()}</span>
              <span>{t('landing.code.footer')}</span>
            </div>
          </div>
        </div>
      </section>

      <section className='landing-features'>
        <div className='landing-container'>
          <div className='landing-features-list'>
            {features.map((f) => (
              <div key={f.no} className='landing-feature'>
                <span className='landing-feature-no'>{f.no}</span>
                <div className='landing-feature-main'>
                  <span className='landing-feature-tag'>{f.tag}</span>
                  <h3 className='landing-feature-title'>{f.title}</h3>
                  <p className='landing-feature-body'>{f.body}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className='landing-cta'>
        <div className='landing-container landing-cta-inner'>
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
        <div className='landing-container landing-footer-inner'>
          <div className='landing-footer-left'>
            <span>
              © {new Date().getFullYear()} {systemName}. All rights reserved.
            </span>
          </div>
          <div className='landing-footer-right'>
            <Link to='/about'>{t('landing.footer.about')}</Link>
            <Link to='/login'>{t('landing.footer.signin')}</Link>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Landing;
