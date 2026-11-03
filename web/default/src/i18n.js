import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import enTranslation from './locales/en/translation.json';
import idTranslation from './locales/id/translation.json';

const supportedLngs = ['en', 'id'];

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    supportedLngs,
    // Keep the stored preference only when it is a language we actually ship.
    // Anything else falls back to English.
    load: 'languageOnly',
    nonExplicitSupportedLngs: true,
    debug: process.env.NODE_ENV === 'development',

    interpolation: {
      escapeValue: false,
    },

    resources: {
      en: {
        translation: enTranslation,
      },
      id: {
        translation: idTranslation,
      },
    },
  });

export default i18n;
