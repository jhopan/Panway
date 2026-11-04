import React from 'react';
import { Card, Icon } from 'semantic-ui-react';
import ChannelsTable from '../../components/ChannelsTable';
import { useTranslation } from 'react-i18next';
import '../Console.css';

const Channel = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='sitemap' />
            {t('channel.title')}
          </Card.Header>
          <div className='console-table-wrap'>
            <ChannelsTable />
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Channel;
