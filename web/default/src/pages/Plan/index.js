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
  Modal,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  isAdmin,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import { renderQuota } from '../../helpers/render';
import '../Console.css';
import '../Dashboard/Dashboard.css';

const emptyDraft = {
  name: '',
  description: '',
  price_idr: '0',
  quota: '0',
  duration_days: '30',
  models: '',
  status: 1,
};

const Plans = () => {
  const { t } = useTranslation();
  const admin = isAdmin();
  const [plans, setPlans] = useState([]);
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(null);
  const [draft, setDraft] = useState(emptyDraft);
  const [saving, setSaving] = useState(false);

  const load = async () => {
    try {
      const [planRes, userRes] = await Promise.all([
        API.get(admin ? '/api/plan/' : '/api/plans'),
        API.get('/api/user/self'),
      ]);
      if (planRes.data.success) {
        setPlans(planRes.data.data || []);
      } else {
        showError(planRes.data.message);
      }
      if (userRes.data.success) {
        setUser(userRes.data.data);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load().then();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const currentPlan = plans.find((p) => user && p.id === user.plan_id);

  const buy = async (plan) => {
    try {
      const res = await API.post(`/api/plans/${plan.id}/buy`);
      if (res.data.success) {
        showSuccess(t('plans.buy_success'));
        await load();
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('plans.buy_failed'));
    }
  };

  const openNew = () => {
    setDraft(emptyDraft);
    setEditing({});
  };

  const openEdit = (plan) => {
    setDraft({
      name: plan.name,
      description: plan.description || '',
      price_idr: String(plan.price_idr),
      quota: String(plan.quota),
      duration_days: String(plan.duration_days),
      models: plan.models || '',
      status: plan.status,
    });
    setEditing(plan);
  };

  const save = async () => {
    setSaving(true);
    const payload = {
      name: draft.name,
      description: draft.description,
      price_idr: parseInt(draft.price_idr, 10) || 0,
      quota: parseInt(draft.quota, 10) || 0,
      duration_days: parseInt(draft.duration_days, 10) || 0,
      models: draft.models,
      status: draft.status,
    };
    try {
      const res = editing.id
        ? await API.put(`/api/plan/${editing.id}`, payload)
        : await API.post('/api/plan/', payload);
      if (res.data.success) {
        showSuccess(t('plans.saved'));
        setEditing(null);
        await load();
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('plans.save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const remove = async (plan) => {
    if (!window.confirm(t('plans.delete_confirm'))) return;
    try {
      const res = await API.delete(`/api/plan/${plan.id}`);
      if (res.data.success) {
        showSuccess(t('plans.deleted'));
        await load();
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('plans.delete_failed'));
    }
  };

  return (
    <div className='dashboard-container'>
      <div className='console-page-head'>
        <div>
          <h2 className='console-page-title'>{t('plans.title')}</h2>
          <p className='console-page-lead'>{t('plans.lead')}</p>
        </div>
        {admin && (
          <Button primary icon='plus' content={t('plans.new')} onClick={openNew} />
        )}
      </div>

      {user && (
        <Card fluid className='chart-card'>
          <Card.Content>
            <Card.Header>
              <Icon name='id badge' /> {t('plans.current')}
            </Card.Header>
            {currentPlan ? (
              <div style={{ marginTop: 8 }}>
                <Label color='green'>{currentPlan.name}</Label>
                <span style={{ marginLeft: 10 }}>
                  {t('plans.expires')}:{' '}
                  {timestamp2string(user.plan_expires_at)}
                </span>
              </div>
            ) : (
              <p style={{ marginTop: 8 }}>{t('plans.no_plan')}</p>
            )}
          </Card.Content>
        </Card>
      )}

      {loading && <div className='catalogue-empty'>{t('plans.loading')}</div>}
      {!loading && plans.length === 0 && (
        <div className='catalogue-empty'>{t('plans.empty')}</div>
      )}

      <div className='catalogue-grid' style={{ marginTop: 20 }}>
        {plans.map((plan) => (
          <div key={plan.id} className='catalogue-card'>
            <div className='catalogue-card-top'>
              <span className='catalogue-card-name'>{plan.name}</span>
              {admin && (
                <span>
                  <button
                    type='button'
                    className='catalogue-edit'
                    onClick={() => openEdit(plan)}
                  >
                    <Icon name='edit' />
                  </button>
                  <button
                    type='button'
                    className='catalogue-edit'
                    onClick={() => remove(plan)}
                  >
                    <Icon name='trash' />
                  </button>
                </span>
              )}
            </div>

            {plan.description && (
              <p style={{ margin: '8px 0', color: 'var(--pw-muted)' }}>
                {plan.description}
              </p>
            )}

            <div className='catalogue-prices'>
              <div className='catalogue-price'>
                <span className='catalogue-price-label'>{t('plans.price')}</span>
                <span className='catalogue-price-value'>
                  {renderQuota(plan.price_idr, t)}
                </span>
              </div>
              <div className='catalogue-price'>
                <span className='catalogue-price-label'>{t('plans.quota')}</span>
                <span className='catalogue-price-value'>
                  {renderQuota(plan.quota, t)}
                </span>
              </div>
            </div>

            <div className='catalogue-card-unit'>
              {t('plans.duration', { days: plan.duration_days })}
              {plan.status === 2 && (
                <Label basic color='grey' style={{ marginLeft: 8 }}>
                  {t('plans.disabled')}
                </Label>
              )}
            </div>
            <div className='catalogue-card-unit'>
              {t('plans.models')}: {plan.models || t('plans.all_models')}
            </div>

            <Button
              primary
              fluid
              style={{ marginTop: 14 }}
              content={t('plans.buy')}
              onClick={() => buy(plan)}
            />
          </div>
        ))}
      </div>

      <Modal open={!!editing} onClose={() => setEditing(null)} size='small'>
        <Modal.Header>
          {editing && editing.id ? t('plans.edit') : t('plans.new')}
        </Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Input
              label={t('plans.name')}
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
            <Form.TextArea
              label={t('plans.description')}
              value={draft.description}
              onChange={(e) =>
                setDraft({ ...draft, description: e.target.value })
              }
            />
            <Form.Group widths='equal'>
              <Form.Field>
                <label>{t('plans.price')}</label>
                <Input
                  type='number'
                  min='0'
                  value={draft.price_idr}
                  onChange={(e) =>
                    setDraft({ ...draft, price_idr: e.target.value })
                  }
                />
              </Form.Field>
              <Form.Field>
                <label>{t('plans.quota')}</label>
                <Input
                  type='number'
                  min='0'
                  value={draft.quota}
                  onChange={(e) =>
                    setDraft({ ...draft, quota: e.target.value })
                  }
                />
              </Form.Field>
              <Form.Field>
                <label>{t('plans.duration_days')}</label>
                <Input
                  type='number'
                  min='1'
                  value={draft.duration_days}
                  onChange={(e) =>
                    setDraft({ ...draft, duration_days: e.target.value })
                  }
                />
              </Form.Field>
            </Form.Group>
            <Form.Input
              label={t('plans.models')}
              placeholder={t('plans.models_hint')}
              value={draft.models}
              onChange={(e) => setDraft({ ...draft, models: e.target.value })}
            />
            <Form.Select
              label={t('plans.status')}
              options={[
                { key: 1, value: 1, text: t('plans.enabled') },
                { key: 2, value: 2, text: t('plans.disabled') },
              ]}
              value={draft.status}
              onChange={(e, { value }) =>
                setDraft({ ...draft, status: value })
              }
            />
            <Message info size='small' content={t('plans.price_hint')} />
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setEditing(null)}>{t('plans.cancel')}</Button>
          <Button primary loading={saving} onClick={save}>
            {t('plans.save')}
          </Button>
        </Modal.Actions>
      </Modal>
    </div>
  );
};

export default Plans;