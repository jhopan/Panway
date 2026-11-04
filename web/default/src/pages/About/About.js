import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Icon } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';
import MarketingShell from '../../components/MarketingShell';
import './About.css';

const About = () => {
  const { t } = useTranslation();
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);
  const [isEmbed, setIsEmbed] = useState(false);

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (data.startsWith('https://')) {
        setIsEmbed(true);
      } else {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout(t('about.loading_failed'));
    }
    setAboutLoaded(true);
  };

  useEffect(() => {
    displayAbout().then();
  }, []);

  // An operator can point "about" at an external page. In that case the
  // content is shown in a frame, exactly like before, but without the
  // console card around it.
  if (isEmbed) {
    return (
      <MarketingShell current='about'>
        <iframe
          src={about}
          className='about-embed'
          title='About'
        />
      </MarketingShell>
    );
  }

  return (
    <MarketingShell current='about'>
      <div className='about'>
        <div className='about-hero'>
          <div className='about-section-inner about-hero-inner'>
            <div className='about-hero-copy'>
              <span className='catalogue-eyebrow'>{t('about.eyebrow')}</span>
              <h1 className='about-title'>{t('about.title')}</h1>
              <p className='about-lead'>{t('about.description')}</p>
            </div>

            <div className='about-hero-panel'>
              <div className='about-hero-panel-bar'>
                <span>{t('about.panel_title')}</span>
              </div>
              <ul className='about-hero-list'>
                {['one_key', 'one_balance', 'one_bill'].map((k) => (
                  <li key={k}>
                    <Icon name='check' />
                    <span>{t(`about.panel.${k}`)}</span>
                  </li>
                ))}
              </ul>
              <div className='about-hero-panel-foot'>
                <code>base_url = https://{t('about.panel_host')}/v1</code>
              </div>
            </div>
          </div>
        </div>

        <div className='about-section'>
          <div className='about-section-inner'>
            <div className='about-grid'>
              <div className='about-card'>
                <span className='about-card-label'>{t('about.what_title')}</span>
                <p className='about-card-body'>{t('about.what_body')}</p>
              </div>
              <div className='about-card'>
                <span className='about-card-label'>
                  {t('about.how_title')}
                </span>
                <p className='about-card-body'>{t('about.how_body')}</p>
              </div>
              <div className='about-card'>
                <span className='about-card-label'>
                  {t('about.pay_title')}
                </span>
                <p className='about-card-body'>{t('about.pay_body')}</p>
              </div>
            </div>
          </div>
        </div>

        {aboutLoaded && about !== '' && (
          <div className='about-section'>
            <div className='about-section-inner'>
              <div className='about-content'
                dangerouslySetInnerHTML={{ __html: about }}
              />
            </div>
          </div>
        )}

        <div className='about-section'>
          <div className='about-section-inner'>
            <div className='about-credits'>
              <div className='about-credit'>
                <span className='about-card-label'>{t('about.credits')}</span>
                <p className='about-card-body'>
                  <a href='https://github.com/jhopan/Panway' target='_blank' rel='noreferrer'>
                    github.com/jhopan/Panway
                  </a>
                </p>
              </div>
              <div className='about-credit'>
                <span className='about-card-label'>
                  {t('about.upstream_title')}
                </span>
                <p className='about-card-body'>
                  {t('about.upstream_body')}
                  <br />
                  <a href='https://github.com/songquanpeng/one-api' target='_blank' rel='noreferrer'>
                    github.com/songquanpeng/one-api
                  </a>
                </p>
              </div>
            </div>

            <div className='about-actions'>
              <Link to='/models' className='landing-btn landing-btn-ghost'>
                {t('landing.nav.models')}
              </Link>
              <Link to='/pricing' className='landing-btn landing-btn-ghost'>
                {t('landing.nav.pricing')}
              </Link>
              <Link to='/login' className='landing-btn landing-btn-primary'>
                {t('landing.nav.signin')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    </MarketingShell>
  );
};

export default About;
