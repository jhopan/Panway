import React from 'react';
import { Card, Icon } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import RedemptionsTable from '../../components/RedemptionsTable';
import '../Console.css';

const Redemption = () => {
  const { t } = useTranslation();
  
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='gift' />
            {t('redemption.title')}
          </Card.Header>
          <div className='console-table-wrap'>
            <RedemptionsTable />
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Redemption;
