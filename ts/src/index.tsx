import { ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';

import './index.css';
import './fonts.css';

declare global {
  interface Window {
    INT_SITE_LOAD: boolean;
    pqcEncrypt: (serverPubKey: string, jsonBody: string) => { blob: string, secret: string };
    pqcDecrypt: (encryptedBlob: string, vaultSecret: string) => string;
  }

  interface IComponent {
    children?: ReactNode;
    loading?: boolean;
    closeModal?(prop?: unknown): void;
  }
}

const rootElement = document.getElementById('root');
if (!rootElement) throw new Error('Failed to find the root element');
const root = createRoot(rootElement);

async function loadInternal() {
  const { store } = (await import('./hooks/store'));
  const { authSlice } = (await import('./hooks/auth'));

  store.dispatch(authSlice.actions.setAuthenticated({ authenticated: true }));

  const App = (await import('./modules/common/App')).default;
  root.render(
    <Provider store={store}>
      <BrowserRouter basename="/app">
        <App />
      </BrowserRouter>
    </Provider>
  );
}

async function loadExternal() {
  const { store } = (await import('./hooks/store'));
  const Ext = (await import('./modules/ext/Ext.tsx')).default;
  root.render(
    <Provider store={store}>
      <BrowserRouter basename="/app/ext">
        <Ext />
      </BrowserRouter>
    </Provider>
  );
}

(async function() {
  try {
    if (window.location.pathname.startsWith('/app/ext/')) {
      await loadExternal();
    } else {
      const currentPathname = window.location.pathname;
      const currentSearch = window.location.search;

      const tzParam = `tz=${Intl.DateTimeFormat().resolvedOptions().timeZone}`

      if (currentPathname.endsWith('/join') && currentSearch.includes('groupCode')) {
        window.location.href = `/auth/register${currentSearch}&${tzParam}`;
        return;
      } else if (currentPathname.endsWith('/register')) {
        window.location.href = `/auth/register?${tzParam}`;
        return;
      }

      const response = await fetch(`/auth/status`, {
        credentials: 'include'
      });

      const loginUrl = `/auth/login?${tzParam}`;
      if (response.ok) {
        const authResponse = (await response.json()) as { authenticated: boolean };
        if (authResponse.authenticated) {
          await loadInternal();
        } else {
          sessionStorage.clear();
          window.location.href = loginUrl;
        }
      } else {
        sessionStorage.clear();
        window.location.href = loginUrl;
      }
    }
  } catch (error) {
    console.error('Application initialization error:', error);
    window.INT_SITE_LOAD = true;
    if (rootElement) {
      root.render(
        <div style={{ fontSize: '1.5rem', padding: '20px', textAlign: 'center', color: 'red' }}>
          An error occurred during application startup. Please try again later.
        </div>
      );
    }
  }
})().catch(err => {
  console.error('Unhandled error in async IIFE:', err);
  window.INT_SITE_LOAD = true;
  if (rootElement) {
    root.render(
      <div style={{ fontSize: '1.5rem', padding: '20px', textAlign: 'center', color: 'red' }}>
        A critical error occurred.
      </div>
    );
  }
});

export { };
