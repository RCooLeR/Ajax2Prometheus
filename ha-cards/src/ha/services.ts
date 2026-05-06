import type { DeviceActionDomain } from '../models/dashboard';
import type { HomeAssistant } from './types';

export async function callEntityService(
  hass: HomeAssistant,
  domain: DeviceActionDomain,
  service: string,
  entityId: string,
): Promise<unknown> {
  try {
    return await hass.callService?.(domain, service, {}, { entity_id: entityId });
  } catch (error) {
    if (!hass.callService) {
      throw error;
    }
    return hass.callService(domain, service, { entity_id: entityId });
  }
}
