import React from 'react';
import { Card, Icon } from 'semantic-ui-react';
import TokensTable from '../../components/TokensTable';
import { useTranslation } from 'react-i18next';
import '../Console.css';

const Token = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='key' />
            {t('token.title')}
          </Card.Header>
          <div className='console-table-wrap'>
            <TokensTable />
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Token;
