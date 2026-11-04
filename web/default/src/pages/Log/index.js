import React from 'react';
import { Card, Icon } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import LogsTable from '../../components/LogsTable';
import '../Console.css';

const Log = () => {
  const { t } = useTranslation();
  
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='history' />
            {t('log.title')}
          </Card.Header>
          <div className='console-table-wrap'>
            <LogsTable />
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Log;
