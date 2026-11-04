import React from 'react';
import { useTranslation } from 'react-i18next';
import { useCatalogue } from './useCatalogue';
import './Catalogue.css';

const Pricing = () => {
  const { t } = useTranslation();
  const { models, loading, failed, formatPrice } = useCatalogue();

  const cheapest = models.reduce(
    (min, m) => (min === null || m.input_price < min.input_price ? m : min),
    null
  );

  return (
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
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={4} className='catalogue-empty'>
                  {t('models.loading')}
                </td>
              </tr>
            )}
            {failed && (
              <tr>
                <td colSpan={4} className='catalogue-empty'>
                  {t('models.error')}
                </td>
              </tr>
            )}
            {!loading && !failed && models.length === 0 && (
              <tr>
                <td colSpan={4} className='catalogue-empty'>
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
  );
};

export default Pricing;
