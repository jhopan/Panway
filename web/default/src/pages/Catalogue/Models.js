import React, { useContext, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Icon, Input, Message, Modal } from 'semantic-ui-react';
import { API, isAdmin, showError, showSuccess } from '../../helpers';
import { UserContext } from '../../context/User';
import { useCatalogue } from './useCatalogue';
import MarketingShell from '../../components/MarketingShell';
import './Catalogue.css';

const Models = () => {
  const { t } = useTranslation();
  const [userState] = useContext(UserContext);
  const { groups, loading, failed, formatPrice } = useCatalogue();

  const admin = isAdmin();

  // The model being edited, and the draft prices inside the modal.
  const [editing, setEditing] = useState(null);
  const [draftInput, setDraftInput] = useState('');
  const [draftOutput, setDraftOutput] = useState('');
  const [saving, setSaving] = useState(false);

  // Local copy of the catalogue so the page reflects a price change
  // immediately instead of waiting for the next fetch.
  const [overrides, setOverrides] = useState({});

  useEffect(() => {
    if (!userState.user) {
      setOverrides({});
    }
  }, [userState.user]);

  const merged = useMemo(() => {
    const result = [];
    groups.forEach(([provider, items]) => {
      const patched = items.map((m) => {
        const override = overrides[m.model_id];
        if (!override) return m;
        return {
          ...m,
          input_price: override.input_price,
          output_price: override.output_price,
        };
      });
      result.push([provider, patched]);
    });
    return result;
  }, [groups, overrides]);

  const openEditor = (model) => {
    setEditing(model);
    setDraftInput(String(Math.round(model.input_price)));
    setDraftOutput(String(Math.round(model.output_price)));
  };

  const closeEditor = () => {
    if (saving) return;
    setEditing(null);
  };

  const savePrice = async () => {
    const inputPrice = parseFloat(draftInput);
    const outputPrice = parseFloat(draftOutput);

    if (!inputPrice || inputPrice <= 0) {
      showError(t('models.invalid_price'));
      return;
    }
    if (!outputPrice || outputPrice <= 0) {
      showError(t('models.invalid_price'));
      return;
    }

    setSaving(true);
    try {
      const res = await API.put('/api/pricing', {
        model_id: editing.model_id,
        input_price: inputPrice,
        output_price: outputPrice,
      });
      if (res.data.success) {
        setOverrides((prev) => ({
          ...prev,
          [editing.model_id]: {
            input_price: inputPrice,
            output_price: outputPrice,
          },
        }));
        showSuccess(t('models.saved'));
        setEditing(null);
      } else {
        showError(res.data.message);
      }
    } catch (err) {
      showError(t('models.save_failed'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <MarketingShell current='models'>
      <div className='catalogue'>
        <div className='catalogue-head'>
          <span className='catalogue-eyebrow'>{t('models.eyebrow')}</span>
          <h1 className='catalogue-title'>{t('models.title')}</h1>
          <p className='catalogue-lead'>{t('models.lead')}</p>
        </div>

        {admin && (
          <Message
            info
            className='catalogue-note'
            icon='pencil'
            header={t('models.admin_title')}
            content={t('models.admin_lead')}
          />
        )}

        {loading && <div className='catalogue-empty'>{t('models.loading')}</div>}
        {failed && <div className='catalogue-empty'>{t('models.error')}</div>}

        {!loading && !failed && merged.length === 0 && (
          <div className='catalogue-empty'>{t('models.empty')}</div>
        )}

        {merged.map(([provider, items]) => (
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
                    {admin && (
                      <button
                        type='button'
                        className='catalogue-edit'
                        onClick={() => openEditor(m)}
                        title={t('models.edit')}
                      >
                        <Icon name='edit' />
                        <span>{t('models.edit')}</span>
                      </button>
                    )}
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

      <Modal open={!!editing} onClose={closeEditor} size='tiny'>
        <Modal.Header>{t('models.modal_title')}</Modal.Header>
        <Modal.Content>
          <div className='catalogue-modal'>
            <code className='catalogue-modal-id'>{editing?.model_id}</code>

            <div className='catalogue-modal-field'>
              <label htmlFor='pw-input-price'>{t('models.input')}</label>
              <Input
                id='pw-input-price'
                type='number'
                min='1'
                step='1000'
                value={draftInput}
                onChange={(e) => setDraftInput(e.target.value)}
                placeholder='50000'
              />
            </div>

            <div className='catalogue-modal-field'>
              <label htmlFor='pw-output-price'>{t('models.output')}</label>
              <Input
                id='pw-output-price'
                type='number'
                min='1'
                step='1000'
                value={draftOutput}
                onChange={(e) => setDraftOutput(e.target.value)}
                placeholder='200000'
              />
            </div>

            <p className='catalogue-modal-hint'>{t('models.modal_hint')}</p>
          </div>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={closeEditor}>{t('models.cancel')}</Button>
          <Button primary loading={saving} onClick={savePrice}>
            {t('models.save')}
          </Button>
        </Modal.Actions>
      </Modal>
    </MarketingShell>
  );
};

export default Models;
