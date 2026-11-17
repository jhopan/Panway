import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Header,
  Icon,
  Input,
  Label,
  Message,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import '../Console.css';
import '../Dashboard/Dashboard.css';

const Security = () => {
  const { t } = useTranslation();
  const [enabled, setEnabled] = useState(false);
  const [setup, setSetup] = useState(null);
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    try {
      const res = await API.get('/api/user/2fa');
      if (res.data.success) {
        setEnabled(!!res.data.data.enabled);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load().then();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const startSetup = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/2fa/setup');
      if (res.data.success) {
        setSetup(res.data.data);
        setCode('');
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('security.invalid'));
    } finally {
      setBusy(false);
    }
  };

  const enable = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/2fa/enable', { code });
      if (res.data.success) {
        showSuccess(t('security.enabled_success'));
        setSetup(null);
        setCode('');
        await load();
      } else {
        showError(res.data.message);
      }
    } finally {
      setBusy(false);
    }
  };

  const disable = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/2fa/disable', { password });
      if (res.data.success) {
        showSuccess(t('security.disabled_success'));
        setPassword('');
        await load();
      } else {
        showError(res.data.message);
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className='dashboard-container'>
      <div className='console-page-head'>
        <div>
          <h2 className='console-page-title'>{t('security.title')}</h2>
          <p className='console-page-lead'>{t('security.lead')}</p>
        </div>
      </div>

      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Icon name='shield alternate' /> {t('security.status')}
          </Card.Header>
          {loading ? (
            <p>{t('security.loading')}</p>
          ) : (
            <div style={{ marginTop: 12 }}>
              <Label color={enabled ? 'green' : 'grey'}>
                {enabled ? t('security.on') : t('security.off')}
              </Label>

              {!enabled && !setup && (
                <>
                  <p style={{ marginTop: 12 }}>{t('security.enable_hint')}</p>
                  <Button primary loading={busy} onClick={startSetup}>
                    {t('security.enable')}
                  </Button>
                </>
              )}

              {!enabled && setup && (
                <div style={{ marginTop: 12 }}>
                  <Message info>
                    <p>{t('security.setup_hint')}</p>
                    <p>
                      <strong>{t('security.secret')}:</strong>{' '}
                      <code>{setup.secret}</code>
                    </p>
                    <p style={{ wordBreak: 'break-all' }}>
                      <strong>{t('security.url')}:</strong>{' '}
                      <code>{setup.url}</code>
                    </p>
                  </Message>
                  <Form onSubmit={enable}>
                    <Form.Input
                      label={t('security.code')}
                      value={code}
                      maxLength={6}
                      onChange={(e) =>
                        setCode(e.target.value.replace(/\D/g, ''))
                      }
                    />
                    <Button primary loading={busy} disabled={code.length !== 6}>
                      {t('security.confirm')}
                    </Button>
                  </Form>
                </div>
              )}

              {enabled && (
                <div style={{ marginTop: 12 }}>
                  <p>{t('security.disable_hint')}</p>
                  <Form onSubmit={disable}>
                    <Form.Field>
                      <label>{t('security.password')}</label>
                      <Input
                        type='password'
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                      />
                    </Form.Field>
                    <Button color='red' loading={busy} disabled={!password}>
                      {t('security.disable')}
                    </Button>
                  </Form>
                </div>
              )}
            </div>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default Security;