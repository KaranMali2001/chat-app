import type { Message } from "./message";

export interface CustomEvent {
  type: string;
  payload: Message;
}

export const EVENT_TYPES = {
  SEND_MESSAGE: "SEND_MESSAGE",
  TYPING: "TYPING",
  MESSAGE_RECEIVED: "MESSAGE_RECEIVED",
};
