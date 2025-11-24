import { useState, useRef, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import Toast from './Toast';

const Chat = ({ isOpen, onClose, chatWith, userID }) => {
  const [messageInput, setMessageInput] = useState('');
  const [chatMessages, setChatMessages] = useState([]);
  const [errorMessage, setErrorMessage] = useState('');
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const { sendMessage, subscribe, connected } = useWebSocketContext();
  const chatWindowRef = useRef(null);
  const inputRef = useRef(null);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [EmojiPicker, setEmojiPicker] = useState(null);
  const [toast, setToast] = useState({ visible: false, message: '', type: 'info' });

  const showToast = (message, type = 'info', duration = 3500) => {
    setToast({ visible: true, message, type, duration });
  };

  const hideToast = () => setToast(prev => ({ ...prev, visible: false }));

  // Fetch chat history when chat opens
  useEffect(() => {
    if (!chatWith || !isOpen) {
      setChatMessages([]); // Clear messages when chat closes
      return;
    }

    const fetchChatHistory = async () => {
      setIsLoadingHistory(true);
      try {
        const response = await fetch(`http://localhost:8080/get-chat-history?user_id=${chatWith.id}`, {
          credentials: 'include',
        });

        if (response.ok) {
          const data = await response.json();
          const messages = data.messages || [];
          // Sort by timestamp (oldest first)
          const sortedMessages = messages.sort((a, b) => 
            new Date(a.timestamp) - new Date(b.timestamp)
          );
          setChatMessages(sortedMessages);
        } else if (response.status === 403) {
          setErrorMessage('You can only chat with users you follow or who follow you');
        } else {
          console.error('Failed to fetch chat history');
        }
      } catch (error) {
        console.error('Error fetching chat history:', error);
      } finally {
        setIsLoadingHistory(false);
      }
    };

    fetchChatHistory();
  }, [chatWith, isOpen]);

  // Filter and listen for messages for this specific chat
  useEffect(() => {
    if (!chatWith || !userID || !isOpen) return;

    const unsubscribe = subscribe((message) => {
      // Handle error messages
      if (message.type === 'error') {
        setErrorMessage(message.content);
        setTimeout(() => setErrorMessage(''), 5000);
        return;
      }

      // Only handle private messages for this specific chat
      if (message.type === 'private_message' || message.type === 'message') {
        // Check if message is between current user and chat partner
        if (
          (message.from === userID && message.to === chatWith.id) || 
          (message.from === chatWith.id && message.to === userID)
        ) {
          setChatMessages(prev => {
            // Avoid duplicates based on timestamp, content, and sender
            const isDuplicate = prev.some(msg => 
              msg.content === message.content &&
              msg.from === message.from &&
              Math.abs(new Date(msg.timestamp) - new Date(message.timestamp)) < 3000
            );
            if (isDuplicate) return prev;
            
            // Add new message and sort by timestamp
            const updated = [...prev, message];
            return updated.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
          });
        }
      }
    });

    return () => unsubscribe();
  }, [chatWith, userID, isOpen, subscribe]);

  // Auto scroll to bottom when new messages arrive
  useEffect(() => {
    if (chatWindowRef.current) {
      chatWindowRef.current.scrollTop = chatWindowRef.current.scrollHeight;
    }
  }, [chatMessages]);

  // Close emoji picker when clicking outside
  useEffect(() => {
    const handler = (e) => {
      if (!showEmojiPicker) return;
      const picker = document.getElementById('emoji-picker');
      const button = document.getElementById('emoji-button');
      if (picker && !picker.contains(e.target) && button && !button.contains(e.target)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('click', handler);
    return () => document.removeEventListener('click', handler);
  }, [showEmojiPicker]);

  // Dynamically load emoji-picker-react on first open to avoid SSR issues
  useEffect(() => {
    let mounted = true;
    if (showEmojiPicker && !EmojiPicker) {
      (async () => {
        try {
          const mod = await import('emoji-picker-react');
          if (mounted) {
            // emoji-picker-react default export is a component
            setEmojiPicker(() => mod.default || mod);
          }
        } catch (err) {
          console.error('Failed to load emoji picker-react:', err);
        }
      })();
    }
    return () => { mounted = false; };
  }, [showEmojiPicker, EmojiPicker]);

  const handleSendMessage = (e) => {
    e.preventDefault();
    
    if (!messageInput.trim() || !chatWith || !connected) return;

    const messageData = {
      type: 'private_message',
      from: userID,
      to: chatWith.id,
      content: messageInput,
      timestamp: new Date().toISOString()
    };

    // Try to send via WebSocket/transport
    const success = sendMessage(messageData);

    if (success) {
      // Optimistically append message so emoji and text appear immediately
      setChatMessages(prev => {
        const updated = [...prev, messageData];
        return updated.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
      });
      setMessageInput('');
    } else {
      showToast('Failed to send message. Please check your connection.', 'error');
    }
  };

  const insertAtCaret = (text) => {
    const input = inputRef.current;
    if (!input) {
      // fallback: append
      setMessageInput(prev => prev + text);
      return;
    }

    // For input/textarea elements
    const start = input.selectionStart ?? input.value.length;
    const end = input.selectionEnd ?? input.value.length;
    const newVal = input.value.slice(0, start) + text + input.value.slice(end);
    setMessageInput(newVal);

    // update caret position after state update
    // using setTimeout to wait for DOM update
    setTimeout(() => {
      const pos = start + text.length;
      input.focus();
      try {
        input.setSelectionRange(pos, pos);
      } catch (err) {
        // ignore
      }
    }, 0);
  };

  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (!isOpen) return null;

  return (
    <div className="group-chat-modern">
      {/* Chat Header */}
      <div className="group-chat-header-modern">
        <div className="group-chat-header-content">
          <div className="group-chat-header-icon">
            {/* show emoji if provided, otherwise initial fallback */}
            <span style={{fontWeight: 700, fontSize: '18px', color: 'white'}}>
              {chatWith?.emoji || (chatWith?.username ? chatWith.username.charAt(0).toUpperCase() : '🙂')}
            </span>
          </div>
          <div className="group-chat-header-info">
            <div className="group-chat-header-title">{chatWith?.username || 'User'}</div>
            <div className="group-chat-status">
              <div className={`group-chat-status-dot ${connected ? '' : 'disconnected'}`}></div>
              <div style={{fontSize: '12px', color: 'rgba(255,255,255,0.9)'}}>{connected ? 'Connected' : 'Disconnected'}</div>
            </div>
          </div>
        </div>
        <div>
          <button onClick={onClose} className="group-chat-close-btn">×</button>
        </div>
      </div>

      {/* Error Message */}
      {errorMessage && (
        <div style={{ 
          backgroundColor: '#fee', 
          border: '1px solid #fcc', 
          color: '#c00', 
          padding: '12px 16px', 
          fontSize: '14px',
          borderRadius: '0'
        }}>
          Warning: {errorMessage}
        </div>
      )}

      {/* Chat Messages */}
      <div ref={chatWindowRef} className="group-chat-messages-modern">
        {isLoadingHistory ? (
          <div style={{textAlign: 'center', color: '#65676b', padding: '32px'}}>
            <p>Loading messages...</p>
          </div>
        ) : chatMessages.length === 0 ? (
          <div style={{textAlign: 'center', color: '#65676b', padding: '32px'}}>
            <p>No messages yet. Start the conversation!</p>
          </div>
        ) : (
          chatMessages.map((msg, index) => {
            const isMyMessage = msg.from === userID;
            return (
              <div key={`${msg.timestamp}-${index}`} className={`group-chat-message-modern ${isMyMessage ? 'my-message' : 'other-message'}`}>
                {!isMyMessage && (
                  <div className="message-username">{msg.username || chatWith?.username || `User ${msg.from}`}</div>
                )}
                <div className="message-bubble-modern">
                  <div className="message-content">{msg.content || msg.message}</div>
                  <div className="message-time">{formatTime(msg.timestamp)}</div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Chat Input */}
      <form onSubmit={handleSendMessage} className="group-chat-input-modern">
        <div className="chat-input-wrapper">
          <button
            id="emoji-button"
            type="button"
            onClick={() => setShowEmojiPicker(s => !s)}
            title="Insert emoji"
            disabled={!connected}
            className="chat-emoji-btn"
          >
            🙂
          </button>
          <input
            type="text"
            ref={inputRef}
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            placeholder={connected ? "Type your message..." : "Connecting..."}
            disabled={!connected}
            className="chat-input-field"
          />
          <button type="submit" disabled={!connected || !messageInput.trim()} className={`chat-send-btn ${connected && messageInput.trim() ? 'active' : ''}`}>
            Send
          </button>
        </div>

        {showEmojiPicker && (
          <div id="emoji-picker" className="chat-emoji-picker-modern">
            {EmojiPicker ? (
              <EmojiPicker
                onEmojiClick={(emojiData) => {
                  const char = emojiData?.emoji || '';
                  insertAtCaret(char);
                  setShowEmojiPicker(false);
                }}
                preload={true}
              />
            ) : (
              <div style={{padding: 12}}>Loading emojis…</div>
            )}
          </div>
        )}
      </form>
        <Toast visible={toast.visible} message={toast.message} type={toast.type} onClose={hideToast} duration={toast.duration} />
    </div>
  );
};

export default Chat;