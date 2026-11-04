import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Icon } from 'semantic-ui-react';
import UsersTable from '../../components/UsersTable';
import '../Console.css';

const User = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header'>
            <Icon name='users' />
            {t('user.title')}
          </Card.Header>
          <div className='console-table-wrap'>
            <UsersTable />
          </div>
        </Card.Content>
      </Card>
    </div>
  );
};

export default User;
