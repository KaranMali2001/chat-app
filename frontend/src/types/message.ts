export interface Message {
  id: string; // Changed from number to string to match Go struct
  sender: string;
  content: string;
  time: string;
  avatar?: string;
  isMe?: boolean; // Frontend-only field for UI
}
