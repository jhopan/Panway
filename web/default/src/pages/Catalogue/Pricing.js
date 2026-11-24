import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import { Button, Icon, Input, Modal } from 'semantic-ui-react';
import { API, isAdmin, showError, showSuccess } from '../../helpers';
import { useCatalogue } from './useCatalogue';
import MarketingShell from '../../components/MarketingShell';
import './Catalogue.css';

const Sparkline = ({ rows }) => {
  if (!rows || rows.length < 2) {
    return <span className='pricing-dim'>—</span>;
  }
  const data = rows.map((row) => ({
    time: row.created_at,
    input: row.input_price,
    output: row.output_price,
  }));
  return (
    <div style={{ width: 110, height: 32 }}>
      <ResponsiveContainer width='100%' height='100%'>
        <LineChart data={data}>
          <Line
            type='monotone'
            dataKey='input'
            stroke='var(--pw-accent, #4f6bff)'
            strokeWidth={2}
            dot={false}
          />
          <Tooltip
            labelFormatter={(value) => new Date(value * 1000).toLocaleString()}
            formatter={(value) => ['Rp' + Math.round(value).toLocaleString(), 'Input']}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

const Pricing = ({ embedded = false }) => {
  const { t } = useTranslation();
  const { models, loading, failed, formatPrice } = useCatalogue();
  const admin = embedded && isAdmin();
  const [history, setHistory] = useState({});
  const [overrides, setOverrides] = useState({});
  const [editing, setEditing] = useState(null);
  const [draftInput, setDraftInput] = useState('');
  const [draftOutput, setDraftOutput] = useState('');
  const [draftCache, setDraftCache] = useState('');
  const [saving, setSaving] = useState(false);

  const loadHistory = () =>
    API.get('/api/pricing/history')
      .then((res) => {
        if (!res.data.success) return;
        const byModel = {};
        (res.data.data || []).forEach((row) => {
          if (!byModel[row.model_id]) byModel[row.model_id] = [];
          byModel[row.model_id].push(row);
        });
        setHistory(byModel);
      })
      .catch(() => {});

  useEffect(() => {
    loadHistory();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const rows = models.map((m) =>
    overrides[m.model_id] ? { ...m, ...overrides[m.model_id] } : m
  );

  const openEditor = (model) => {
    setEditing(model);
    setDraftInput(String(Math.round(model.input_price)));
    setDraftOutput(String(Math.round(model.output_price)));
    setDraftCache(String(Math.round(model.cache_hit_price || 0)));
  };

  const savePrice = async () => {
    const inputPrice = parseFloat(draftInput);
    const outputPrice = parseFloat(draftOutput);
    const cachePrice = parseFloat(draftCache) || 0;
    if (!inputPrice || inputPrice <= 0 || !outputPrice || outputPrice <= 0) {
      showError(t('models.invalid_price'));
      return;
    }
    if (cachePrice < 0 || cachePrice > inputPrice) {
      showError(t('models.invalid_price'));
      return;
    }
    setSaving(true);
    try {
      const res = await API.put('/api/pricing', {
        model_id: editing.model_id,
        input_price: inputPrice,
        output_price: outputPrice,
        cache_hit_price: cachePrice,
      });
      if (res.data.success) {
        setOverrides((prev) => ({
          ...prev,
          [editing.model_id]: {
            input_price: inputPrice,
            output_price: outputPrice,
            cache_hit_price: cachePrice,
          },
        }));
        showSuccess(t('models.saved'));
        setEditing(null);
        await loadHistory();
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('models.save_failed'));
    } finally {
      setSaving(false);
    }
  };

  const cheapest = rows.reduce(
    (min, m) => (min === null || m.input_price < min.input_price ? m : min),
    null
  );

  const page = (
    <>
      <div className='catalogue'>
      <div className='catalogue-head'>
        <span className='catalogue-eyebrow'>{t('pricing.eyebrow')}</span>
        <h1 className='catalogue-title'>{t('pricing.title')}</h1>
        <p className='catalogue-lead'>{t('pricing.lead')}</p>
      </div>

      {!loading && !failed && cheapest && (
        <div className='pricing-highlight'>
          <span className='pricing-highlight-label'>
            {t('pricing.cheapest')}
          </span>
          <span className='pricing-highlight-value'>
            {formatPrice(cheapest.input_price)}
          </span>
          <span className='pricing-highlight-unit'>{t('models.unit')}</span>
          <code className='pricing-highlight-model'>{cheapest.model_id}</code>
        </div>
      )}

      <div className='pricing-table-wrap'>
        <table className='pricing-table'>
          <thead>
            <tr>
              <th>{t('models.model')}</th>
              <th>{t('models.input')}</th>
              <th>{t('models.output')}</th>
              <th>{t('pricing.cache_hit')}</th>
              <th>{t('pricing.multiplier')}</th>
              <th>{t('pricing.calls_7d')}</th>
              <th>{t('pricing.avg_latency')}</th>
              <th>{t('pricing.history')}</th>
              {admin && <th>{t('pricing.actions')}</th>}
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={admin ? 9 : 8} className='catalogue-empty'>
                  {t('models.loading')}
                </td>
              </tr>
            )}
            {failed && (
              <tr>
                <td colSpan={admin ? 9 : 8} className='catalogue-empty'>
                  {t('models.error')}
                </td>
              </tr>
            )}
            {!loading && !failed && rows.length === 0 && (
              <tr>
                <td colSpan={admin ? 9 : 8} className='catalogue-empty'>
                  {t('models.empty')}
                </td>
              </tr>
            )}
            {rows.map((m) => (
              <tr key={m.model_id}>
                <td>
                  <code>{m.model_id}</code>
                </td>
                <td>{formatPrice(m.input_price)}</td>
                <td>{formatPrice(m.output_price)}</td>
                <td>{formatPrice(m.cache_hit_price || 0)}</td>
                <td className='pricing-dim'>
                  {m.output_price / m.input_price}x
                </td>
                <td>{(m.requests_7d || 0).toLocaleString()}</td>
                <td>
                  {m.avg_latency_ms ? `${m.avg_latency_ms} ms` : '—'}
                </td>
                <td>
                  <Sparkline rows={history[m.model_id]} />
                </td>
                {admin && (
                  <td>
                    <Button
                      basic
                      size='mini'
                      icon='edit'
                      content={t('models.edit')}
                      onClick={() => openEditor(m)}
                    />
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <p className='pricing-disclaimer'>{t('pricing.disclaimer')}</p>

      <div className='pricing-notes'>
        <div className='pricing-note'>
          <span className='pricing-note-title'>{t('pricing.note_billing')}</span>
          <span className='pricing-note-body'>{t('pricing.note_billing_body')}</span>
        </div>
        <div className='pricing-note'>
          <span className='pricing-note-title'>{t('pricing.note_topup')}</span>
          <span className='pricing-note-body'>{t('pricing.note_topup_body')}</span>
        </div>
        <div className='pricing-note'>
          <span className='pricing-note-title'>{t('pricing.note_rate')}</span>
          <span className='pricing-note-body'>{t('pricing.note_rate_body')}</span>
        </div>
      </div>
      </div>

      <Modal
        open={!!editing}
        onClose={() => {
          if (!saving) setEditing(null);
        }}
        size='tiny'
      >
        <Modal.Header>{t('models.modal_title')}</Modal.Header>
        <Modal.Content>
          <div className='catalogue-modal'>
            <code className='catalogue-modal-id'>{editing?.model_id}</code>

            <div className='catalogue-modal-field'>
              <label htmlFor='pricing-input'>{t('models.input')}</label>
              <Input
                id='pricing-input'
                type='number'
                min='1'
                step='1000'
                value={draftInput}
                onChange={(e) => setDraftInput(e.target.value)}
                placeholder='50000'
              />
            </div>

            <div className='catalogue-modal-field'>
              <label htmlFor='pricing-output'>{t('models.output')}</label>
              <Input
                id='pricing-output'
                type='number'
                min='1'
                step='1000'
                value={draftOutput}
                onChange={(e) => setDraftOutput(e.target.value)}
                placeholder='200000'
              />
            </div>

            <div className='catalogue-modal-field'>
              <label htmlFor='pricing-cache'>{t('pricing.cache_hit')}</label>
              <Input
                id='pricing-cache'
                type='number'
                min='0'
                step='1000'
                value={draftCache}
                onChange={(e) => setDraftCache(e.target.value)}
                placeholder='0'
              />
            </div>

            <p className='catalogue-modal-hint'>{t('models.modal_hint')}</p>
          </div>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setEditing(null)}>{t('models.cancel')}</Button>
          <Button primary loading={saving} onClick={savePrice}>
            {t('models.save')}
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );

  if (embedded) {
    return page;
  }

  return <MarketingShell current='pricing'>{page}</MarketingShell>;
};

export default Pricing;
