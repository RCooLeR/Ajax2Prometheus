export interface HomeAssistantState {
  entity_id?: string;
  state: string;
  attributes: Record<string, unknown>;
  last_changed?: string;
  last_updated?: string;
}

export interface HomeAssistant {
  states: Record<string, HomeAssistantState>;
  callWS?<TResponse = unknown>(message: {
    type: string;
    [key: string]: unknown;
  }): Promise<TResponse>;
  user?: unknown;
  themes?: unknown;
  locale?: unknown;
}

export interface AjaxLovelaceDetailedCardConfig {
  type: string;
  default_room?: string;
  account?: string;
}

export interface AjaxLovelaceChipsCardConfig {
  type: string;
  max_chips?: number;
  account?: string;
}
