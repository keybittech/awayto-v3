import React, { ComponentType, lazy, ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';

import './index.css';
import './fonts.css';

import { store } from 'awayto/hooks';

declare global {
  interface Window {
    INT_SITE_LOAD: boolean;
  }

  interface IProps {
    children?: ReactNode;
    loading?: boolean;
    closeModal?(prop?: unknown): void;
  }

  interface IComponent {
    children?: ReactNode;
    loading?: boolean;
    closeModal?(prop?: unknown): void;
  }
}

const root = createRoot(document.getElementById('root') as Element);

if (window.location.pathname.startsWith('/app/ext/')) {
  (async function() {
    try {
      const Ext = (await import('./modules/ext/Ext.tsx')).default;
      root.render(
        <Provider store={store}>
          <BrowserRouter basename="/app/ext">
            <Ext />
          </BrowserRouter>
        </Provider>
      );
    } catch (err) {
      const error = err as Error
      console.log('error loading kiosk', error);
    }
  })().catch(console.error);
} else {
  (async function() {
    try {
      const AuthProvider = (await import('./modules/auth/AuthProvider.tsx')).default;
      root.render(
        <Provider store={store}>
          <BrowserRouter basename="/app">
            <AuthProvider />
          </BrowserRouter>
        </Provider>
      );
    } catch (error) {
      console.log('the final error', error)
    }
  })().catch(console.error);
}
