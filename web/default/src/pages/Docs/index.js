import React, { useEffect, useState } from 'react';
import { Card, Header, Icon, Label, Message } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { API } from '../../helpers';
import '../Console.css';
import '../Dashboard/Dashboard.css';
import './Docs.css';

const Docs = () => {
  const { t } = useTranslation();
  const [health, setHealth] = useState(null);

  const loadHealth = async () => {
    try {
      const res = await API.get('/healthz');
      setHealth(res.data);
    } catch (err) {
      setHealth({
        status: 'unreachable',
        database: 'unknown',
        channels: 0,
        version: '',
      });
    }
  };

  useEffect(() => {
    loadHealth().then();
    const timer = setInterval(loadHealth, 30000);
    return () => clearInterval(timer);
  }, []);

  const baseUrl = window.location.origin + '/v1';
  const example = `curl ${baseUrl}/chat/completions \\
  -H "Authorization: Bearer sk-YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "agnes/agnes-3.0-flash",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`;

  return (
    <div className='dashboard-container'>
      <div className='console-page-head'>
        <div>
          <h2 className='console-page-title'>{t('docs.title')}</h2>
          <p className='console-page-lead'>{t('docs.lead')}</p>
        </div>
      </div>

      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Icon name='heartbeat' /> {t('docs.status')}
          </Card.Header>
          <div style={{ marginTop: 12, display: 'flex', gap: 16, flexWrap: 'wrap' }}>
            <span>
              {t('docs.service')}:{' '}
              <Label color={health?.status === 'ok' ? 'green' : 'red'}>
                {health?.status || '...'}
              </Label>
            </span>
            <span>
              {t('docs.database')}:{' '}
              <Label color={health?.database === 'ok' ? 'green' : 'red'}>
                {health?.database || '...'}
              </Label>
            </span>
            <span>
              {t('docs.channels')}: <strong>{health?.channels ?? '...'}</strong>
            </span>
            <span>
              {t('docs.version')}: <code>{health?.version || '...'}</code>
            </span>
          </div>
          <p style={{ marginTop: 12 }}>
            <a href='/healthz' target='_blank' rel='noreferrer'>
              /healthz
            </a>
          </p>
        </Card.Content>
      </Card>

      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Icon name='code' /> {t('docs.quickstart')}
          </Card.Header>
          <Message info>
            <p>
              {t('docs.base_url')}: <code>{baseUrl}</code>
            </p>
            <p>
              {t('docs.auth')}:{' '}
              <code>Authorization: Bearer sk-YOUR_TOKEN</code>
            </p>
          </Message>
          <Header as='h4'>{t('docs.chat_endpoint')}</Header>
          <pre className='docs-code'>
            <code>{example}</code>
          </pre>
          <Header as='h4'>{t('docs.other_endpoints')}</Header>
          <ul>
            <li>
              <code>GET /v1/models</code> — {t('docs.models_hint')}
            </li>
            <li>
              <code>GET /api/pricing</code> — {t('docs.pricing_hint')}
            </li>
            <li>
              <code>POST /v1/embeddings</code>, <code>/v1/images/generations</code>
            </li>
          </ul>
          <Message
            warning
            content={t('docs.rate_limit_hint')}
            icon='info circle'
          />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Docs;