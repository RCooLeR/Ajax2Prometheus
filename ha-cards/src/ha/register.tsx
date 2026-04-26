import type { ReactNode } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { DetailedDashboardView } from '../views/DetailedDashboardView';
import { ChipsOverviewView } from '../views/ChipsOverviewView';
import { setAssetBaseUrl } from '../utils/assets';
import type {
  AjaxLovelaceChipsCardConfig,
  AjaxLovelaceDetailedCardConfig,
  HomeAssistant,
} from './types';
import themeCss from '../styles/theme.css?raw';
import dashboardCss from '../styles/dashboard.css?raw';

setAssetBaseUrl(new URL(/* @vite-ignore */ './assets/', import.meta.url).toString());
const cardCss = `${themeCss}\n${dashboardCss}`;

abstract class ReactHomeAssistantElement<TConfig> extends HTMLElement {
  protected root: Root | null = null;
  protected mountNode: HTMLDivElement | null = null;
  protected hassValue?: HomeAssistant;
  protected configValue?: TConfig;

  connectedCallback() {
    this.style.display = 'block';

    if (!this.root) {
      const shadowRoot = this.shadowRoot ?? this.attachShadow({ mode: 'open' });
      shadowRoot.replaceChildren();

      const style = document.createElement('style');
      style.textContent = cardCss;
      shadowRoot.appendChild(style);

      const mountNode = document.createElement('div');
      mountNode.className = 'ajax-lovelace-shadow-root';
      shadowRoot.appendChild(mountNode);

      this.mountNode = mountNode;
      this.root = createRoot(mountNode);
    }

    this.renderReact();
  }

  disconnectedCallback() {
    this.root?.unmount();
    this.root = null;
    this.mountNode = null;
  }

  set hass(hass: HomeAssistant) {
    this.hassValue = hass;
    this.renderReact();
  }

  protected renderReact() {
    if (!this.root) {
      return;
    }

    this.root.render(this.renderNode());
  }

  protected abstract renderNode(): ReactNode;
}

class AjaxLovelaceDetailedCard extends ReactHomeAssistantElement<AjaxLovelaceDetailedCardConfig> {
  setConfig(config: AjaxLovelaceDetailedCardConfig) {
    this.configValue = config;
    this.style.height = '100%';
    this.renderReact();
  }

  getCardSize() {
    return 20;
  }

  getGridOptions() {
    return {
      columns: 'full',
      min_rows: 10,
      rows: 12,
    };
  }

  protected renderNode() {
    const config = this.configValue ?? { type: 'custom:ajax-lovelace-detailed-card' };

    return (
      <div className="ajax-ha-card ajax-ha-card--detailed">
        <DetailedDashboardView
          mode="embedded"
          initialRoomId={config.default_room}
          hass={this.hassValue}
          account={config.account}
        />
      </div>
    );
  }

  static getStubConfig(): Omit<AjaxLovelaceDetailedCardConfig, 'type'> {
    return {
      default_room: 'living-room',
    };
  }
}

class AjaxLovelaceChipsCard extends ReactHomeAssistantElement<AjaxLovelaceChipsCardConfig> {
  setConfig(config: AjaxLovelaceChipsCardConfig) {
    if (config.max_chips !== undefined && config.max_chips < 1) {
      throw new Error('max_chips must be greater than 0');
    }

    this.configValue = config;
    this.renderReact();
  }

  getCardSize() {
    return 3;
  }

  getGridOptions() {
    return {
      columns: 'full',
      rows: 3,
      min_rows: 2,
      max_rows: 4,
    };
  }

  protected renderNode() {
    const config = this.configValue ?? { type: 'custom:ajax-lovelace-chips-card' };

    return (
      <div className="ajax-ha-card ajax-ha-card--chips">
        <ChipsOverviewView maxChips={config.max_chips} hass={this.hassValue} account={config.account} />
      </div>
    );
  }

  static getStubConfig(): Omit<AjaxLovelaceChipsCardConfig, 'type'> {
    return {
      max_chips: 7,
    };
  }
}

if (!customElements.get('ajax-lovelace-detailed-card')) {
  customElements.define('ajax-lovelace-detailed-card', AjaxLovelaceDetailedCard);
}

if (!customElements.get('ajax-lovelace-chips-card')) {
  customElements.define('ajax-lovelace-chips-card', AjaxLovelaceChipsCard);
}

window.customCards = window.customCards || [];
window.customCards.push(
  {
    type: 'ajax-lovelace-detailed-card',
    name: 'Ajax Lovelace Detailed',
    description: 'Fullscreen Ajax security dashboard card for panel views.',
    preview: false,
  },
  {
    type: 'ajax-lovelace-chips-card',
    name: 'Ajax Lovelace Chips',
    description: 'Compact Ajax overview card with global state chips.',
    preview: false,
  },
);

declare global {
  interface Window {
    customCards?: Array<{
      type: string;
      name: string;
      description?: string;
      preview?: boolean;
    }>;
  }
}
