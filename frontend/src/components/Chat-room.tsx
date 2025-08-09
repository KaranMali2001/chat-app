import { AlertCircle, Loader2, MessageCircle, Send, Users } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';

const ChatApp = () => {
  // State management
  const [username, setUsername] = useState('');
  const [roomId, setRoomId] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [connectedUsers, setConnectedUsers] = useState(new Set());
  const [error, setError] = useState('');
  const [roomStats, setRoomStats] = useState({
    active_rooms: 0,
    total_connections: 0,
  });
  const [isCreatingRoom, setIsCreatingRoom] = useState(false);

  // Refs
  const socketRef = useRef(null);
  const messagesEndRef = useRef(null);
  const baseUrl = 'localhost:80/api/v1'; // Change this to your backend URL

  // Scroll to bottom when new messages arrive
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Fetch room statistics
  const fetchRoomStats = useCallback(async () => {
    try {
      const response = await fetch(`http://${baseUrl}/room-stats`);
      if (response.ok) {
        const stats = await response.json();
        setRoomStats(stats);
      }
    } catch (err) {
      console.error('Failed to fetch room stats:', err);
    }
  }, [baseUrl]);

  useEffect(() => {
    fetchRoomStats();
    const interval = setInterval(fetchRoomStats, 30000); // Update every 30 seconds
    return () => clearInterval(interval);
  }, [fetchRoomStats]);

  // Create a new room
  const createRoom = async () => {
    if (!username.trim()) {
      setError('Please enter a username');
      return;
    }

    setIsCreatingRoom(true);
    setError('');

    try {
      const response = await fetch(`http://${baseUrl}/create-room?username=${encodeURIComponent(username.trim())}`, { method: 'POST' });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();

      if (data.success) {
        setRoomId(data.message.data);
        await connectToRoom(data.message.data, username.trim());
      } else {
        setError('Failed to create room');
      }
    } catch (err) {
      setError(`Failed to create room: ${err.message}`);
    } finally {
      setIsCreatingRoom(false);
    }
  };

  // Connect to existing room
  const joinRoom = async () => {
    if (!username.trim() || !roomId.trim()) {
      setError('Please enter both username and room ID');
      return;
    }

    await connectToRoom(roomId.trim(), username.trim());
  };

  // WebSocket connection logic
  const connectToRoom = async (room, user) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.close();
    }

    setIsConnecting(true);
    setError('');

    try {
      const wsUrl = `ws://${baseUrl}/ws?username=${encodeURIComponent(user)}&roomid=${encodeURIComponent(room)}`;
      const socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setIsConnecting(false);
        setConnectedUsers(new Set([user]));
      };

      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          handleWebSocketMessage(message);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Connection failed. Please check your connection and try again.');
        setIsConnecting(false);
      };

      socket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        setIsConnecting(false);

        if (event.code !== 1000) {
          // Not a normal closure
          setError('Connection lost. Please try reconnecting.');
        }
      };

      socketRef.current = socket;
    } catch (err) {
      setError(`Connection failed: ${err.message}`);
      setIsConnecting(false);
    }
  };

  // Handle incoming WebSocket messages
  const handleWebSocketMessage = (message) => {
    switch (message.type) {
      case 'message_received':
        setMessages((prev) => [
          ...prev,
          {
            ...message.payload,
            timestamp: new Date(message.payload.time),
            type: 'message',
          },
        ]);
        break;

      case 'user_joined':
        setConnectedUsers((prev) => new Set([...prev, message.payload.sender]));
        setMessages((prev) => [
          ...prev,
          {
            id: `join-${Date.now()}`,
            content: `${message.payload.sender} joined the room`,
            timestamp: new Date(message.payload.time),
            type: 'system',
            sender: 'System',
          },
        ]);
        break;

      case 'user_left':
        setConnectedUsers((prev) => {
          const newSet = new Set(prev);
          newSet.delete(message.payload.sender);
          return newSet;
        });
        setMessages((prev) => [
          ...prev,
          {
            id: `leave-${Date.now()}`,
            content: `${message.payload.sender} left the room`,
            timestamp: new Date(message.payload.time),
            type: 'system',
            sender: 'System',
          },
        ]);
        break;

      case 'error':
        setError(message.payload.content);
        break;

      default:
        console.log('Unknown message type:', message.type);
    }
  };

  // Send message
  const sendMessage = () => {
    if (!newMessage.trim() || !socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      return;
    }

    const message = {
      type: 'send_message',
      payload: {
        id: crypto.randomUUID(),
        sender: username,
        content: newMessage.trim(),
        room_id: roomId,
        time: new Date().toISOString(),
      },
    };

    socketRef.current.send(JSON.stringify(message));
    setNewMessage('');
  };

  // Handle Enter key press
  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // Disconnect from room
  const disconnect = () => {
    if (socketRef.current) {
      const leaveMessage = {
        type: 'leave_room',
        payload: {
          room_id: roomId,
          sender: username,
          time: new Date().toISOString(),
        },
      };

      if (socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.send(JSON.stringify(leaveMessage));
      }

      socketRef.current.close();
    }

    setIsConnected(false);
    setMessages([]);
    setConnectedUsers(new Set());
    setRoomId('');
    setError('');
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, []);

  // Format timestamp
  const formatTime = (date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (!isConnected) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg max-w-md w-full p-8">
          <div className="text-center mb-8">
            <MessageCircle className="mx-auto h-12 w-12 text-indigo-600 mb-4" />
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Join Chat Room</h1>
            <div className="text-sm text-gray-500 space-y-1">
              <p>Active Rooms: {roomStats.active_rooms}</p>
              <p>Total Users: {roomStats.total_connections}</p>
            </div>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
              <AlertCircle className="h-4 w-4 flex-shrink-0" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
                placeholder="Enter your username"
                disabled={isConnecting || isCreatingRoom}
              />
            </div>

            <div className="space-y-3">
              <button
                onClick={createRoom}
                disabled={isConnecting || isCreatingRoom || !username.trim()}
                className="w-full bg-indigo-600 text-white py-2 px-4 rounded-lg hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isCreatingRoom ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Creating Room...
                  </>
                ) : (
                  'Create New Room'
                )}
              </button>

              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-300" />
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-2 bg-white text-gray-500">or</span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Room ID</label>
                <input
                  type="text"
                  value={roomId}
                  onChange={(e) => setRoomId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
                  placeholder="Enter room ID"
                  disabled={isConnecting || isCreatingRoom}
                />
              </div>

              <button
                onClick={joinRoom}
                disabled={isConnecting || isCreatingRoom || !username.trim() || !roomId.trim()}
                className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isConnecting ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Connecting...
                  </>
                ) : (
                  'Join Room'
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen bg-gray-100 flex flex-col">
      {/* Header */}
      <div className="bg-white shadow-sm border-b px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MessageCircle className="h-6 w-6 text-indigo-600" />
          <div>
            <h1 className="font-semibold text-gray-900">Room: {roomId}</h1>
            <p className="text-sm text-gray-500">Connected as {username}</p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <Users className="h-4 w-4" />
            <span>{connectedUsers.size} online</span>
          </div>
          <button onClick={disconnect} className="px-3 py-1 text-sm bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors">
            Leave Room
          </button>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <MessageCircle className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>No messages yet. Start the conversation!</p>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.type === 'system' ? 'justify-center' : message.sender === username ? 'justify-end' : 'justify-start'}`}
            >
              {message.type === 'system' ? (
                <div className="text-xs text-gray-500 bg-gray-100 px-3 py-1 rounded-full">{message.content}</div>
              ) : (
                <div
                  className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                    message.sender === username ? 'bg-indigo-600 text-white' : 'bg-white text-gray-900 shadow-sm border'
                  }`}
                >
                  {message.sender !== username && <div className="text-xs font-medium text-gray-600 mb-1">{message.sender}</div>}
                  <div className="break-words">{message.content}</div>
                  <div className={`text-xs mt-1 ${message.sender === username ? 'text-indigo-200' : 'text-gray-500'}`}>
                    {formatTime(message.timestamp)}
                  </div>
                </div>
              )}
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Message Input */}
      <div className="bg-white border-t p-4">
        <div className="flex gap-2">
          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type your message..."
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
            disabled={!isConnected}
          />
          <button
            onClick={sendMessage}
            disabled={!newMessage.trim() || !isConnected}
            className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChatApp;
