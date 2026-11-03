import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Icon, Message, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';

// TelegramVerification lets a customer prove they own a Telegram chat.
//
// It is a verification, not a login: the customer signs in with email or
// Google first, then confirms a code sent to their chat.
const TelegramVerification = () => {
  const { t } = useTranslation();
  const [status, setStatus] = useState(null);
  const [chatId, setChatId] = useState('');
  const [code, setCode] = useState('');
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/user/telegram/status');
      if (res.data.success) {
        setStatus(res.data.data);
        setChatId(res.data.data.chat_id || '');
      }
    } catch (error) {
      // The status endpoint is optional; ignore failures.
    }
  };

  useEffect(() => {
    loadStatus();
  }, []);

  const sendCode = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/telegram/send', { chat_id: chatId });
      const { success, message, data } = res.data;
      if (success) {
        setSent(true);
        showSuccess(t('user.telegram.code_sent'));
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('user.telegram.request_failed'));
    } finally {
      setBusy(false);
    }
  };

  const verify = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/telegram/verify', {
        chat_id: chatId,
        code: code,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('user.telegram.verified'));
        setCode('');
        setSent(false);
        loadStatus();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('user.telegram.request_failed'));
    } finally {
      setBusy(false);
    }
  };

  const unverify = async () => {
    setBusy(true);
    try {
      const res = await API.post('/api/user/telegram/unverify', {});
      if (res.data.success) {
        showSuccess(t('user.telegram.unverified'));
        loadStatus();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('user.telegram.request_failed'));
    } finally {
      setBusy(false);
    }
  };

  if (!status || !status.enabled) {
    return null;
  }

  if (status.verified) {
    return (
      <Segment>
        <Header as='h4'>
          <Icon name='telegram' />
          {t('user.telegram.title')}
        </Header>
        <Message positive size='small' content={t('user.telegram.already_verified')} />
        <Form>
          <Form.Field>
            <Form.Input
              label={t('user.telegram.chat_id')}
              value={chatId}
              readOnly
            />
          </Form.Field>
          <Button basic color='red' onClick={unverify} loading={busy} disabled={busy}>
            {t('user.telegram.unverify_button')}
          </Button>
        </Form>
      </Segment>
    );
  }

  return (
    <Segment>
      <Header as='h4'>
        <Icon name='telegram' />
        {t('user.telegram.title')}
      </Header>
      <Message info size='small' content={t('user.telegram.hint', { bot: status.bot_username })} />
      <Form>
        <Form.Group widths='equal'>
          <Form.Input
            label={t('user.telegram.chat_id')}
            placeholder='123456789'
            value={chatId}
            onChange={(e) => setChatId(e.target.value)}
          />
          <Form.Field>
            <Button
              type='button'
              onClick={sendCode}
              loading={busy}
              disabled={busy || !chatId}
              style={{ marginTop: '1.7em' }}
            >
              {t('user.telegram.send_button')}
            </Button>
          </Form.Field>
        </Form.Group>
        {sent && (
          <Form.Group widths='equal'>
            <Form.Input
              label={t('user.telegram.code')}
              placeholder='123456'
              maxLength={6}
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
            <Form.Field>
              <Button
                type='button'
                primary
                onClick={verify}
                loading={busy}
                disabled={busy || code.length !== 6}
                style={{ marginTop: '1.7em' }}
              >
                {t('user.telegram.verify_button')}
              </Button>
            </Form.Field>
          </Form.Group>
        )}
      </Form>
    </Segment>
  );
};

export default TelegramVerification;
