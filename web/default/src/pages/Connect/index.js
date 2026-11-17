import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { copy, showSuccess, showWarning } from '../../helpers';
import './Connect.css';

const CLIENTS = [
  'Cursor',
  'Continue',
  'Aider',
  'Open WebUI',
  'n8n',
  'LangChain',
  'Dify',
  'LobeChat',
  'Cherry Studio',
  'AnythingLLM',
];

const CodeBlock = ({ label, code }) => {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const onCopy = async () => {
    if (await copy(code)) {
      setCopied(true);
      showSuccess(t('connect.copied'));
      setTimeout(() => setCopied(false), 1500);
    } else {
      showWarning(t('connect.copy_failed'));
    }
  };

  return (
    <div className='connect-block'>
      <div className='connect-block-head'>
        <span>{label}</span>
        <button type='button' className='connect-copy' onClick={onCopy}>
          {copied ? t('connect.copied') : t('connect.copy')}
        </button>
      </div>
      <pre className='connect-code'>
        <code>{code}</code>
      </pre>
    </div>
  );
};

const Connect = () => {
  const { t } = useTranslation();
  const origin = window.location.origin;
  const baseUrl = origin + '/v1';

  const envSnippet = `export OPENAI_BASE_URL="${baseUrl}"
export OPENAI_API_KEY="sk-YOUR_TOKEN"`;

  const cliSnippet = `panway connect --client aider \\
  --base-url ${baseUrl} \\
  --key sk-YOUR_TOKEN`;

  const curlSnippet = `curl ${baseUrl}/chat/completions \\
  -H "Authorization: Bearer sk-YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "MODEL_ID",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`;

  const pythonSnippet = `from openai import OpenAI

client = OpenAI(
    base_url="${baseUrl}",
    api_key="sk-YOUR_TOKEN",
)

reply = client.chat.completions.create(
    model="MODEL_ID",
    messages=[{"role": "user", "content": "Hello"}],
)
print(reply.choices[0].message.content)`;

  const nodeSnippet = `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${baseUrl}",
  apiKey: "sk-YOUR_TOKEN",
});

const reply = await client.chat.completions.create({
  model: "MODEL_ID",
  messages: [{ role: "user", content: "Hello" }],
});
console.log(reply.choices[0].message.content);`;

  return (
    <div className='connect'>
        <div className='connect-hero'>
          <span className='connect-eyebrow'>{t('connect.eyebrow')}</span>
          <h1 className='connect-title'>{t('connect.title')}</h1>
          <p className='connect-lead'>{t('connect.lead')}</p>
          <div className='connect-actions'>
            <Link to='/token' className='landing-btn landing-btn-primary'>
              {t('connect.get_key')}
            </Link>
            <a
              className='landing-btn landing-btn-ghost'
              href={baseUrl + '/models'}
              target='_blank'
              rel='noreferrer'
            >
              {t('connect.models')}
            </a>
          </div>
        </div>

        <div className='connect-section'>
          <div className='connect-steps'>
            <div className='connect-step'>
              <span className='connect-step-no'>1</span>
              <h3>{t('connect.step1_title')}</h3>
              <p>{t('connect.step1_body')}</p>
            </div>
            <div className='connect-step'>
              <span className='connect-step-no'>2</span>
              <h3>{t('connect.step2_title')}</h3>
              <p>{t('connect.step2_body')}</p>
            </div>
            <div className='connect-step'>
              <span className='connect-step-no'>3</span>
              <h3>{t('connect.step3_title')}</h3>
              <p>{t('connect.step3_body')}</p>
            </div>
          </div>
        </div>

        <div className='connect-section'>
          <h2 className='connect-section-title'>{t('connect.snippets')}</h2>
          <CodeBlock label={t('connect.cli')} code={cliSnippet} />
          <CodeBlock label={t('connect.env')} code={envSnippet} />
          <CodeBlock label='curl' code={curlSnippet} />
          <CodeBlock label='Python (openai)' code={pythonSnippet} />
          <CodeBlock label='Node (openai)' code={nodeSnippet} />
        </div>

        <div className='connect-section'>
          <h2 className='connect-section-title'>{t('connect.clients')}</h2>
          <p className='connect-section-lead'>{t('connect.clients_lead')}</p>
          <div className='connect-clients'>
            {CLIENTS.map((c) => (
              <span key={c} className='connect-client'>
                {c}
              </span>
            ))}
          </div>
        </div>

        <div className='connect-section'>
          <h2 className='connect-section-title'>{t('connect.safety')}</h2>
          <ul className='connect-safety'>
            <li>{t('connect.safety1')}</li>
            <li>{t('connect.safety2')}</li>
            <li>{t('connect.safety3')}</li>
            <li>{t('connect.safety4')}</li>
          </ul>
        </div>
      </div>
  );
};

export default Connect;