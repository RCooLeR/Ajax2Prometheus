import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { setAssetBaseUrl } from './utils/assets';
import './styles/theme.css';
import './styles/dashboard.css';
import './styles/standalone.css';

setAssetBaseUrl(new URL(/* @vite-ignore */ './', import.meta.url).toString());

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
