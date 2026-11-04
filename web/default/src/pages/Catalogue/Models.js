import React from 'react';
import { useTranslation } from 'react-i18next';
import { useCatalogue } from './useCatalogue';
import MarketingShell from '../../components/MarketingShell';
import './Catalogue.css';

const Models = () => {
  const { t } = useTranslation();
  const { groups, loading, failed, formatPrice } = useCatalogue();

  return (
    <MarketingShell current='models'>
      <div className='catalogue'>
      <div className='catalogue-head'>
        <span className='catalogue-eyebrow'>{t('models.eyebrow')}</span>
        <h1 className='catalogue-title'>{t('models.title')}</h1>
        <p className='catalogue-lead'>{t('models.lead')}</p>
      </div>

      {loading && <div className='catalogue-empty'>{t('models.loading')}</div>}
      {failed && <div className='catalogue-empty'>{t('models.error')}</div>}

      {!loading && !failed && groups.length === 0 && (
        <div className='catalogue-empty'>{t('models.empty')}</div>
      )}

      {groups.map(([provider, items]) => (
        <div key={provider || 'other'} className='catalogue-group'>
          <div className='catalogue-group-head'>
            <span className='catalogue-group-name'>
              {provider ? provider.toUpperCase() : t('models.other')}
            </span>
            <span className='catalogue-group-count'>
              {items.length}{' '}
              {items.length === 1 ? t('models.model') : t('models.models')}
            </span>
          </div>

          <div className='catalogue-grid'>
            {items.map((m) => (
              <div key={m.model_id} className='catalogue-card'>
                <div className='catalogue-card-top'>
                  {m.provider && (
                    <span className='catalogue-card-provider'>
                      {m.provider.toUpperCase()}
                    </span>
                  )}
                  <span className='catalogue-card-name'>{m.name}</span>
                </div>

                <div className='catalogue-prices'>
                  <div className='catalogue-price'>
                    <span className='catalogue-price-label'>
                      {t('models.input')}
                    </span>
                    <span className='catalogue-price-value'>
                      {formatPrice(m.input_price)}
                    </span>
                  </div>
                  <div className='catalogue-price'>
                    <span className='catalogue-price-label'>
                      {t('models.output')}
                    </span>
                    <span className='catalogue-price-value'>
                      {formatPrice(m.output_price)}
                    </span>
                  </div>
                </div>

                <div className='catalogue-card-unit'>{t('models.unit')}</div>

                <code className='catalogue-card-id'>{m.model_id}</code>
              </div>
            ))}
          </div>
        </div>
      ))}
      </div>
    </MarketingShell>
  );
};

export default Models;
