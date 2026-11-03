import React from 'react';
import { useLocation } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import Header from './Header';
import Footer from './Footer';
import Sidebar from './Sidebar';
import App from '../App';
import { ToastContainer } from 'react-toastify';

// The landing page renders its own marketing chrome.
const BARE_ROUTES = ['/'];

// Auth screens keep the simple top bar: a sidebar there would be noise
// because there is nothing to navigate to yet.
const PLAIN_ROUTES = [
  '/login',
  '/register',
  '/reset',
  '/oauth/github',
  '/oauth/lark',
];

const Chrome = () => {
  const location = useLocation();
  const path = location.pathname;

  if (BARE_ROUTES.includes(path)) {
    return (
      <>
        <App />
        <ToastContainer />
      </>
    );
  }

  if (PLAIN_ROUTES.some((r) => path.startsWith(r))) {
    return (
      <div className='plain-shell'>
        <Header />
        <Container className={'main-content'}>
          <App />
        </Container>
        <ToastContainer />
        <Footer />
      </div>
    );
  }

  return (
    <div className='console-shell'>
      <Sidebar />
      <main className='console-main'>
        <Container className={'main-content'}>
          <App />
        </Container>
      </main>
      <ToastContainer />
    </div>
  );
};

export default Chrome;
