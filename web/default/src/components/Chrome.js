import React from 'react';
import { useLocation } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import Header from './Header';
import Footer from './Footer';
import App from '../App';
import { ToastContainer } from 'react-toastify';

// Routes that render their own marketing chrome (header, footer, spacing).
const BARE_ROUTES = ['/'];

const Chrome = () => {
  const location = useLocation();
  const isBare = BARE_ROUTES.includes(location.pathname);

  if (isBare) {
    return (
      <>
        <App />
        <ToastContainer />
      </>
    );
  }

  return (
    <>
      <Header />
      <Container className={'main-content'}>
        <App />
      </Container>
      <ToastContainer />
      <Footer />
    </>
  );
};

export default Chrome;
