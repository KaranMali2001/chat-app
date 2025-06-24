"use client";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { format } from "date-fns";

import type { CustomEvent } from "@/types/event";
import { EVENT_TYPES } from "@/types/event";
import type { Message } from "@/types/message";
import { MoreVertical, Phone, Send, Video } from "lucide-react";
import { useEffect, useState } from "react";

export type EventType = (typeof EVENT_TYPES)[keyof typeof EVENT_TYPES];

export default function ChatScreen() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);

  useEffect(() => {
    console.log(
      "Connecting to WebSocket server...",
      import.meta.env.VITE_WEBSOCKET_URL
    );
    const socket = new WebSocket(import.meta.env.VITE_WEBSOCKET_URL);

    socket.onopen = () => {
      console.log("socket connected");
      setIsConnected(true);
      setSocket(socket);
    };

    socket.onclose = () => {
      console.log("socket disconnected");
      setIsConnected(false);
    };

    socket.onmessage = (event) => {
      console.log("Message received:", event.data);
      try {
        const parsedEvent = JSON.parse(event.data) as CustomEvent;
        console.log("Parsed Event:", parsedEvent);

        switch (parsedEvent.type) {
          case EVENT_TYPES.MESSAGE_RECEIVED:
          case EVENT_TYPES.SEND_MESSAGE:
            console.log("Received message:", parsedEvent.payload);
            const newMessage: Message = {
              ...parsedEvent.payload,
              time: parsedEvent.payload.time || format(Date.now(), "hh:mm a"),
              isMe: parsedEvent.payload.sender === "You", // Determine if it's from current user
            };
            setMessages((prev) => [...prev, newMessage]);
            break;

          case EVENT_TYPES.TYPING:
            console.log("Typing event:", parsedEvent.payload);
            // For typing events, you might want to handle differently
            // This assumes the payload contains typing info in the content field
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

  const sendEvent = (eventType: EventType, message: Message) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      console.error("WebSocket is not connected");
      return;
    }

    const event: CustomEvent = {
      type: eventType,
      payload: message,
    };

    console.log("Sending event:", event);
    socket.send(JSON.stringify(event));
  };

  const handleSendMessage = () => {
    if (!input.trim() || !socket || socket.readyState !== WebSocket.OPEN)
      return;

    const message: Message = {
      id: Date.now().toString(), // Convert to string to match Go struct
      sender: "You",
      content: input,
      time: format(Date.now(), "hh:mm a"),
      isMe: true,
    };

    // Send the event to the backend
    sendEvent(EVENT_TYPES.SEND_MESSAGE, message);

    // Add to local messages immediately for better UX
    setMessages((prev) => [...prev, message]);
    setInput("");
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);

    // Send typing event (you can customize this based on your backend needs)
    if (e.target.value.length > 0 && !isTyping) {
      const typingMessage: Message = {
        id: `typing-${Date.now()}`,
        sender: "You",
        content: "true", // Indicate typing started
        time: format(Date.now(), "hh:mm a"),
      };
      // sendEvent(EVENT_TYPES.TYPING, typingMessage);
    } else if (e.target.value.length === 0 && isTyping) {
      const typingMessage: Message = {
        id: `typing-${Date.now()}`,
        sender: "You",
        content: "false", // Indicate typing stopped
        time: format(Date.now(), "hh:mm a"),
      };
      // sendEvent(EVENT_TYPES.TYPING, typingMessage);
    }
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
          <Button
            variant="ghost"
            size="icon"
            className="text-gray-400 hover:text-white"
          >
            <Phone className="h-5 w-5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="text-gray-400 hover:text-white"
          >
            <Video className="h-5 w-5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="text-gray-400 hover:text-white"
          >
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
                    src={
                      message.avatar || "/placeholder.svg?height=32&width=32"
                    }
                    alt={message.sender}
                  />
                  <AvatarFallback>
                    {message.sender.charAt(0).toUpperCase()}
                  </AvatarFallback>
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

        {/* Typing indicator */}
        {isTyping && (
          <div className="flex justify-start">
            <div className="flex items-end space-x-2 max-w-xs lg:max-w-md">
              <Avatar className="h-8 w-8">
                <AvatarImage
                  src="/placeholder.svg?height=32&width=32"
                  alt="Alice Johnson"
                />
                <AvatarFallback>AJ</AvatarFallback>
              </Avatar>
              <div className="px-4 py-2 rounded-2xl bg-gray-700 text-white rounded-bl-sm border border-gray-600">
                <div className="flex items-center space-x-1">
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                    <div
                      className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: "0.1s" }}
                    ></div>
                    <div
                      className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"
                      style={{ animationDelay: "0.2s" }}
                    ></div>
                  </div>
                  <p className="text-sm text-gray-300 ml-2">typing...</p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className="bg-gray-800 border-t border-gray-700 p-4">
        <div className="flex items-center space-x-2">
          <Input
            value={input}
            onChange={handleInputChange}
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
            disabled={!isConnected}
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
