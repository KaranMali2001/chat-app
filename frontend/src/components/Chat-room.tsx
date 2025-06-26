"use client";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { CustomEvent } from "@/types/event";
import { EVENT_TYPES } from "@/types/event";
import type { Message } from "@/types/message";
import { format } from "date-fns";
import { MoreVertical, Phone, Send, Video } from "lucide-react";
import { useEffect, useState } from "react";

export default function ChatScreenDarkForest() {
  const [username, setUsername] = useState<string>("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);

  useEffect(() => {
    const u = window.prompt("Enter your username");
    if (!u) return;
    setUsername(u);

    const socket = new WebSocket(
      `${import.meta.env.VITE_WEBSOCKET_URL}?username=${u}`
    );

    socket.onopen = () => {
      console.log("Socket connected");
      setIsConnected(true);
      setSocket(socket);
    };

    socket.onclose = () => {
      console.log("Socket disconnected");
      setIsConnected(false);
    };

    socket.onmessage = (event) => {
      try {
        const parsedEvent = JSON.parse(event.data) as CustomEvent;
        switch (parsedEvent.type) {
          case EVENT_TYPES.MESSAGE_RECEIVED:
          case EVENT_TYPES.SEND_MESSAGE:
          case EVENT_TYPES.BROADCAST:
            const newMessage: Message = {
              ...parsedEvent.payload,
              time: parsedEvent.payload.time || format(Date.now(), "hh:mm a"),
              isMe: parsedEvent.payload.sender === username,
            };
            setMessages((prev) => [...prev, newMessage]);
            break;

          case EVENT_TYPES.TYPING:
            setIsTyping(parsedEvent.payload.content === "true");
            break;

          default:
            console.error("Unknown event type:", parsedEvent.type);
            break;
        }
      } catch (error) {
        console.error("Error parsing message:", error);
      }
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    return () => {
      socket.close();
    };
  }, []);

  const sendEvent = (eventType: string, message: Message) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.error("WebSocket is not connected");
      return;
    }

    const event: CustomEvent = {
      type: eventType,
      payload: message,
    };

    socket.send(JSON.stringify(event));
  };

  const handleSendMessage = () => {
    if (!input.trim()) return;

    const message: Message = {
      id: Date.now().toString(),
      sender: username,
      content: input,
      time: format(Date.now(), "hh:mm a"),
      isMe: true,
    };

    sendEvent(EVENT_TYPES.BROADCAST, message);
    setMessages((prev) => [...prev, message]);
    setInput("");
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
  };

  return (
    <div className="flex flex-col h-screen bg-zinc-900">
      {/* Header */}
      <div className="bg-emerald-950 border-b border-emerald-900/50 px-4 py-3 flex items-center justify-between shadow-lg">
        <div className="flex items-center space-x-3">
          <Avatar className="h-10 w-10 ring-2 ring-emerald-600/40">
            <AvatarImage
              src="/placeholder.svg?height=40&width=40"
              alt={username}
            />
            <AvatarFallback className="bg-emerald-700 text-emerald-100">
              {username.slice(0, 2).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div>
            <h2 className="font-semibold text-emerald-100">{username}</h2>
            <p className="text-sm text-teal-400 font-medium">
              {isConnected ? "Online" : "Offline"}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="icon"
            className="text-emerald-300 hover:text-emerald-200 hover:bg-emerald-900/50"
          >
            <Phone className="h-5 w-5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="text-emerald-300 hover:text-emerald-200 hover:bg-emerald-900/50"
          >
            <Video className="h-5 w-5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="text-emerald-300 hover:text-emerald-200 hover:bg-emerald-900/50"
          >
            <MoreVertical className="h-5 w-5" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex flex-col space-y-2 ${
              message.isMe ? "items-end" : "items-start"
            }`}
          >
            {!message.isMe && (
              <p className="text-sm text-emerald-200 font-medium px-3">
                {message.sender}
              </p>
            )}
            <div
              className={`flex items-end space-x-3 max-w-xs lg:max-w-md ${
                message.isMe ? "flex-row-reverse space-x-reverse" : ""
              }`}
            >
              {!message.isMe && (
                <Avatar className="h-8 w-8 ring-2 ring-zinc-700">
                  <AvatarImage
                    src={
                      message.avatar || "/placeholder.svg?height=32&width=32"
                    }
                    alt={message.sender}
                  />
                  <AvatarFallback className="bg-zinc-700 text-zinc-300 text-xs">
                    {message.sender.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              )}
              <div
                className={`px-4 py-3 rounded-2xl shadow-lg ${
                  message.isMe
                    ? "bg-gradient-to-r from-emerald-700 to-teal-700 text-emerald-50 rounded-br-md"
                    : "bg-zinc-800 text-zinc-100 rounded-bl-md border border-zinc-700"
                }`}
              >
                <p className="text-sm leading-relaxed">{message.content}</p>
                <p
                  className={`text-xs mt-2 ${
                    message.isMe ? "text-emerald-200" : "text-zinc-400"
                  }`}
                >
                  {message.time}
                </p>
              </div>
            </div>
          </div>
        ))}

        {isTyping && (
          <div className="flex justify-start">
            <div className="flex items-end space-x-3 max-w-xs lg:max-w-md">
              <Avatar className="h-8 w-8 ring-2 ring-zinc-700">
                <AvatarImage
                  src="/placeholder.svg?height=32&width=32"
                  alt="Someone"
                />
                <AvatarFallback className="bg-zinc-700 text-zinc-300 text-xs">
                  ...
                </AvatarFallback>
              </Avatar>
              <div className="px-4 py-3 rounded-2xl bg-zinc-800 text-zinc-100 rounded-bl-md border border-zinc-700 shadow-lg">
                <div className="flex items-center space-x-2">
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce"></div>
                    <div
                      className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce"
                      style={{ animationDelay: "0.1s" }}
                    ></div>
                    <div
                      className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce"
                      style={{ animationDelay: "0.2s" }}
                    ></div>
                  </div>
                  <p className="text-sm text-zinc-400">typing...</p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Input */}
      <div className="bg-emerald-950 border-t border-emerald-900/50 p-4 shadow-lg">
        <div className="flex items-center space-x-3">
          <Input
            value={input}
            onChange={handleInputChange}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleSendMessage();
            }}
            placeholder="Type a message..."
            className="flex-1 rounded-full border-emerald-800 bg-zinc-800 text-emerald-100 placeholder:text-zinc-400 focus:border-emerald-600 focus:ring-emerald-600 focus:bg-zinc-700"
          />
          <Button
            size="icon"
            onClick={handleSendMessage}
            className="rounded-full bg-gradient-to-r from-emerald-700 to-teal-700 hover:from-emerald-800 hover:to-teal-800 shadow-lg"
            disabled={!isConnected}
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
