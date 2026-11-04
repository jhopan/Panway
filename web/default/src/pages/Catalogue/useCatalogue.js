import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API } from '../../helpers';
import './Catalogue.css';

// Shared by the Models page and the Pricing page: one fetch, one shape.
export const useCatalogue = () => {
  const { i18n } = useTranslation();
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    API.get('/api/pricing')
      .then((res) => {
        const body = res.data;
        if (body.success && body.data && Array.isArray(body.data.models)) {
          setModels(body.data.models);
        } else {
          setFailed(true);
        }
      })
      .catch(() => setFailed(true))
      .finally(() => setLoading(false));
  }, []);

  const locale = i18n.language === 'id' ? 'id-ID' : 'en-US';

  const formatPrice = useMemo(
    () => (value) => 'Rp' + Math.round(value).toLocaleString(locale),
    [locale]
  );

  // Group models by provider so the page reads as a catalogue, not a list.
  const groups = useMemo(() => {
    const map = new Map();
    models.forEach((m) => {
      const key = m.provider || '';
      if (!map.has(key)) {
        map.set(key, []);
      }
      map.get(key).push(m);
    });
    return [...map.entries()];
  }, [models]);

  return { models, groups, loading, failed, formatPrice };
};
