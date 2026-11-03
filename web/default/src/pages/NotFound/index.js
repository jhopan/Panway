import React from 'react';
import { Message } from 'semantic-ui-react';

const NotFound = () => (
  <>
    <Message negative>
      <Message.Header>Page not found</Message.Header>
      <p>Please check whether the address in your browser is correct</p>
    </Message>
  </>
);

export default NotFound;
