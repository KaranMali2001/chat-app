export interface Message {
  id: string;
  sender: string;
  content: string;
  time: string;
  avatar?: string;
  isMe?: boolean;
}
