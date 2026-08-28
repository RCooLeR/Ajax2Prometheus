import type { DetailedHTMLProps, HTMLAttributes } from 'react';
import type { HomeAssistant, HomeAssistantState } from './ha/types';

type HomeAssistantCameraStreamElement = HTMLElement & {
  hass?: HomeAssistant;
  stateObj?: HomeAssistantState;
};

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'ha-camera-stream': DetailedHTMLProps<
        HTMLAttributes<HomeAssistantCameraStreamElement>,
        HomeAssistantCameraStreamElement
      >;
    }
  }
}
