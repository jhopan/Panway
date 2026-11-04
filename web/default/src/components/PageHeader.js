import React from 'react';
import { Icon } from 'semantic-ui-react';

// A consistent page heading for the console pages: title, optional lead
// text and an icon, so every console screen opens the same way.
const PageHeader = ({ title, lead, icon }) => {
  return (
    <div className='console-page-head'>
      <div className='console-page-head-main'>
        <h1 className='console-page-title'>
          {icon && <Icon name={icon} />}
          {title}
        </h1>
        {lead && <p className='console-page-lead'>{lead}</p>}
      </div>
    </div>
  );
};

export default PageHeader;
