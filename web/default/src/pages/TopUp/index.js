import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Card,
  Statistic,
  Message,
  Icon,
  Label,
} from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess } from '../../helpers';
import { renderQuota } from '../../helpers/render';
import { useTranslation } from 'react-i18next';

const TopUp = () => {
  const { t } = useTranslation();
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [user, setUser] = useState({});

  // PanWallet
  const [wallet, setWallet] = useState(null);
  const [walletId, setWalletId] = useState('');
  const [walletBusy, setWalletBusy] = useState(false);
  const [syncAmount, setSyncAmount] = useState('');

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo(t('topup.redeem_code.empty_code'));
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode,
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('topup.redeem_code.success'));
        setUserQuota((quota) => quota + data);
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('topup.redeem_code.request_failed'));
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError(t('topup.redeem_code.no_link'));
      return;
    }
    let url = new URL(topUpLink);
    url.searchParams.append('username', user.username);
    url.searchParams.append('user_id', user.id);
    url.searchParams.append('transaction_id', crypto.randomUUID());
    window.open(url.toString(), '_blank');
  };

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setUserQuota(data.quota);
      setUser(data);
    } else {
      showError(message);
    }
  };

  // Muat wallet yang sudah tertaut.
  const loadWallet = async () => {
    try {
      const res = await API.get('/api/user/wallet');
      if (res.data.success && res.data.data) {
        setWallet(res.data.data);
        setWalletId(res.data.data.public_id || '');
      }
    } catch (err) {
      // Belum tertaut — form tetap kosong.
    }
  };

  const linkWallet = async () => {
    const id = walletId.trim();
    if (!id) {
      showInfo(t('topup.wallet.id_required'));
      return;
    }
    setWalletBusy(true);
    try {
      const res = await API.post('/api/user/wallet/link', { public_id: id });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('topup.wallet.linked'));
        setWallet(data);
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('topup.wallet.request_failed'));
    } finally {
      setWalletBusy(false);
    }
  };

  const unlinkWallet = async () => {
    setWalletBusy(true);
    try {
      const res = await API.post('/api/user/wallet/unlink', {});
      if (res.data.success) {
        setWallet(null);
        setWalletId('');
        showSuccess(t('topup.wallet.unlinked'));
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('topup.wallet.request_failed'));
    } finally {
      setWalletBusy(false);
    }
  };

  // Sinkronkan saldo PanWallet menjadi kuota.
  const syncWallet = async () => {
    const amount = parseInt(syncAmount, 10);
    if (!amount || amount <= 0) {
      showInfo(t('topup.wallet.amount_required'));
      return;
    }
    setWalletBusy(true);
    try {
      const res = await API.post('/api/user/wallet/sync', {
        amount_idr: amount,
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('topup.wallet.synced', { quota: data.quota }));
        setUserQuota((quota) => quota + data.quota);
        setSyncAmount('');
        await loadWallet();
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('topup.wallet.request_failed'));
    } finally {
      setWalletBusy(false);
    }
  };

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    getUserQuota().then();
    loadWallet().then();
  }, []);

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>
            <Header as='h2'>{t('topup.title')}</Header>
          </Card.Header>

          <Message
            info
            icon={<Icon name='wallet' />}
            header={t('topup.wallet.title')}
            content={t('topup.wallet.description')}
          />

          {wallet ? (
            <Card fluid style={{ marginTop: '1em' }}>
              <Card.Content>
                <Grid stackable columns={3}>
                  <Grid.Column>
                    <Statistic size='small' label={t('topup.wallet.public_id')}>
                      <Statistic.Value>
                        <Label color='blue' size='large'>
                          {wallet.public_id}
                        </Label>
                      </Statistic.Value>
                    </Statistic>
                  </Grid.Column>
                  <Grid.Column>
                    <Statistic size='small' label={t('topup.wallet.balance')}>
                      <Statistic.Value>
                        Rp {(wallet.balance ?? 0).toLocaleString('id-ID')}
                      </Statistic.Value>
                      <Statistic.Label>{wallet.currency || 'IDR'}</Statistic.Label>
                    </Statistic>
                  </Grid.Column>
                  <Grid.Column>
                    <Statistic size='small' label={t('topup.wallet.available')}>
                      <Statistic.Value>
                        Rp {(wallet.available ?? 0).toLocaleString('id-ID')}
                      </Statistic.Value>
                    </Statistic>
                  </Grid.Column>
                </Grid>

                <Form style={{ marginTop: '1em' }}>
                  <Form.Group widths='equal'>
                    <Form.Input
                      fluid
                      label={t('topup.wallet.sync_amount')}
                      placeholder='25000'
                      type='number'
                      min='1'
                      value={syncAmount}
                      onChange={(e) => setSyncAmount(e.target.value)}
                    />
                    <Form.Field>
                      <Button
                        primary
                        fluid
                        loading={walletBusy}
                        disabled={walletBusy}
                        type='button'
                        onClick={syncWallet}
                        style={{ marginTop: '1.7em' }}
                      >
                        {t('topup.wallet.sync_button')}
                      </Button>
                    </Form.Field>
                  </Form.Group>
                  <Button
                    type='button'
                    basic
                    color='red'
                    loading={walletBusy}
                    disabled={walletBusy}
                    onClick={unlinkWallet}
                  >
                    {t('topup.wallet.unlink_button')}
                  </Button>
                </Form>
              </Card.Content>
            </Card>
          ) : (
            <Card fluid style={{ marginTop: '1em' }}>
              <Card.Content>
                <Form>
                  <Form.Group widths='equal'>
                    <Form.Input
                      fluid
                      label={t('topup.wallet.id_label')}
                      placeholder='JZXJFADE'
                      value={walletId}
                      onChange={(e) => setWalletId(e.target.value)}
                    />
                    <Form.Field>
                      <Button
                        primary
                        fluid
                        loading={walletBusy}
                        disabled={walletBusy}
                        type='button'
                        onClick={linkWallet}
                        style={{ marginTop: '1.7em' }}
                      >
                        {t('topup.wallet.link_button')}
                      </Button>
                    </Form.Field>
                  </Form.Group>
                </Form>
              </Card.Content>
            </Card>
          )}

          <Grid columns={2} stackable style={{ marginTop: '2em' }}>
            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content
                  style={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <Card.Header>
                    <Header as='h3' style={{ color: '#2185d0', margin: '1em' }}>
                      <i className='credit card icon'></i>
                      {t('topup.get_code.title')}
                    </Header>
                  </Card.Header>
                  <Card.Description
                    style={{
                      flex: 1,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      style={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'space-between',
                      }}
                    >
                      <div style={{ textAlign: 'center', paddingTop: '1em' }}>
                        <Statistic>
                          <Statistic.Value style={{ color: '#2185d0' }}>
                            {renderQuota(userQuota, t)}
                          </Statistic.Value>
                          <Statistic.Label>
                            {t('topup.get_code.current_quota')}
                          </Statistic.Label>
                        </Statistic>
                      </div>

                      <div style={{ textAlign: 'center', paddingBottom: '1em' }}>
                        <Button
                          primary
                          size='large'
                          onClick={openTopUpLink}
                          style={{ width: '80%' }}
                        >
                          {t('topup.get_code.button')}
                        </Button>
                      </div>
                    </div>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>

            <Grid.Column>
              <Card
                fluid
                style={{
                  height: '100%',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              >
                <Card.Content
                  style={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                  }}
                >
                  <Card.Header>
                    <Header as='h3' style={{ color: '#21ba45', margin: '1em' }}>
                      <i className='ticket alternate icon'></i>
                      {t('topup.redeem_code.title')}
                    </Header>
                  </Card.Header>
                  <Card.Description
                    style={{
                      flex: 1,
                      display: 'flex',
                      flexDirection: 'column',
                    }}
                  >
                    <div
                      style={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Form.Input
                        fluid
                        icon='key'
                        iconPosition='left'
                        placeholder={t('topup.redeem_code.placeholder')}
                        value={redemptionCode}
                        onChange={(e) => {
                          setRedemptionCode(e.target.value);
                        }}
                        onPaste={(e) => {
                          e.preventDefault();
                          const pastedText = e.clipboardData.getData('text');
                          setRedemptionCode(pastedText.trim());
                        }}
                        action={
                          <Button
                            icon='paste'
                            content={t('topup.redeem_code.paste')}
                            onClick={async () => {
                              try {
                                const text =
                                  await navigator.clipboard.readText();
                                setRedemptionCode(text.trim());
                              } catch (err) {
                                showError(t('topup.redeem_code.paste_error'));
                              }
                            }}
                          />
                        }
                      />

                      <div style={{ paddingBottom: '1em' }}>
                        <Button
                          color='green'
                          fluid
                          size='large'
                          onClick={topUp}
                          loading={isSubmitting}
                          disabled={isSubmitting}
                        >
                          {isSubmitting
                            ? t('topup.redeem_code.submitting')
                            : t('topup.redeem_code.submit')}
                        </Button>
                      </div>
                    </div>
                  </Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid>
        </Card.Content>
      </Card>
    </div>
  );
};

export default TopUp;
