import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import { API } from '../../helpers';
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
  const [history, setHistory] = useState({});

  useEffect(() => {
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
  }, []);

  const cheapest = models.reduce(
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
              <th>{t('pricing.multiplier')}</th>
              <th>{t('pricing.calls_7d')}</th>
              <th>{t('pricing.avg_latency')}</th>
              <th>{t('pricing.history')}</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={7} className='catalogue-empty'>
                  {t('models.loading')}
                </td>
              </tr>
            )}
            {failed && (
              <tr>
                <td colSpan={7} className='catalogue-empty'>
                  {t('models.error')}
                </td>
              </tr>
            )}
            {!loading && !failed && models.length === 0 && (
              <tr>
                <td colSpan={7} className='catalogue-empty'>
                  {t('models.empty')}
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
    </>
  );

  if (embedded) {
    return page;
  }

  return <MarketingShell current='pricing'>{page}</MarketingShell>;
};

export default Pricing;
