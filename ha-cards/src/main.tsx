import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { setAssetBaseUrl } from './utils/assets';
import './styles/theme.css';
import './styles/dashboard.css';
import './styles/standalone.css';

setAssetBaseUrl(new URL(/* @vite-ignore */ './', import.meta.url).toString());

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Unable to mount AjaxBridge: #root was not found');
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
