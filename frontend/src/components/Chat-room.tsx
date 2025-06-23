"use client";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { format } from "date-fns";

import { MoreVertical, Phone, Send, Video } from "lucide-react";
import { useEffect, useState } from "react";
export interface Message {
  id: number;
  sender: string;
  content: string;
  time: string;
  isMe: boolean;
  avatar?: string;
}

export default function ChatScreen() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [input, setInput] = useState("");
  useEffect(() => {
    const socket = new WebSocket(import.meta.env.WEBSOCKET_URL);
    socket.onopen = () => {
      console.log("socket connected");
      setIsConnected(true);
      setSocket(socket);
    };
    socket.onclose = () => {
      console.log("socket disconnected");
    };
    socket.onmessage = (event) => {
      console.log("Message", event);
      // const message = JSON.parse(event.data) as Message;
      const message: Message = {
        id: Date.now(),
        sender: "SERVER",
        content: event.data,
        time: format(Date.now(), "hh:mm a"),
        isMe: false,
      };
      setMessages((prev) => [...prev, message]);
    };
    socket.onerror = () => {
      console.error("error");
    };

    return () => {
      socket.close();
    };
  }, []);
  const handleSendMessage = () => {
    if (!input.trim() || !socket || socket.readyState !== WebSocket.OPEN)
      return;
    const message: Message = {
      id: Date.now(),
      sender: "You",
      content: input,
      time: format(Date.now(), "hh:mm a"),
      isMe: true,
    };
    socket.send(JSON.stringify(message));
    setMessages((prev) => [...prev, message]);
  };
  console.log("is connected", isConnected);
  return (
    <div className="flex flex-col h-screen bg-gray-900">
      {/* Header */}
      <div className="bg-gray-800 border-b border-gray-700 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <Avatar className="h-10 w-10">
            <AvatarImage
              src="/placeholder.svg?height=40&width=40"
              alt="Alice Johnson"
            />
            <AvatarFallback>AJ</AvatarFallback>
          </Avatar>
          <div>
            <h2 className="font-semibold text-white">Alice Johnson</h2>
            <p className="text-sm text-green-400">
              {isConnected ? "Online" : "Offline"}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="ghost" size="icon">
            <Phone className="h-5 w-5" />
          </Button>
          <Button variant="ghost" size="icon">
            <Video className="h-5 w-5" />
          </Button>
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-5 w-5" />
          </Button>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${message.isMe ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`flex items-end space-x-2 max-w-xs lg:max-w-md ${
                message.isMe ? "flex-row-reverse space-x-reverse" : ""
              }`}
            >
              {!message.isMe && (
                <Avatar className="h-8 w-8">
                  <AvatarImage
                    src={message.avatar || "/placeholder.svg"}
                    alt={message.sender}
                  />
                  <AvatarFallback>AJ</AvatarFallback>
                </Avatar>
              )}
              <div
                className={`px-4 py-2 rounded-2xl ${
                  message.isMe
                    ? "bg-blue-500 text-white rounded-br-sm"
                    : "bg-gray-700 text-white rounded-bl-sm border border-gray-600"
                }`}
              >
                <p className="text-sm">{message.content}</p>
                <p
                  className={`text-xs mt-1 ${
                    message.isMe ? "text-blue-100" : "text-gray-300"
                  }`}
                >
                  {message.time}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Input Area */}
      <div className="bg-gray-800 border-t border-gray-700 p-4">
        <div className="flex items-center space-x-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleSendMessage();
            }}
            placeholder="Type a message..."
            className="flex-1 rounded-full border-gray-600 bg-gray-700 text-white placeholder:text-gray-400 focus:border-blue-500 focus:ring-blue-500"
          />
          <Button
            size="icon"
            onClick={handleSendMessage}
            className="rounded-full bg-blue-500 hover:bg-blue-600"
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
