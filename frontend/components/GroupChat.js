import { useState, useEffect, useRef } from 'react';

export default function GroupChat({ groupId, user, isOpen, onClose, connected, messages, sendMessage }) {
  const [messageInput, setMessageInput] = useState('');
  const [groupMessages, setGroupMessages] = useState([]);
  const [existingMessages, setExistingMessages] = useState([]);
  const chatWindowRef = useRef(null);
  const inputRef = useRef(null);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [EmojiPicker, setEmojiPicker] = useState(null);

  const userId = user?.id || user?.userID || user?.user_id;

  // Load existing messages when chat opens
  useEffect(() => {
    if (isOpen && groupId) {
      console.log('🔄 Loading messages for group:', groupId);
      fetch(`/api/group-messages?group_id=${groupId}`, { credentials: 'include' })
        .then(res => {
          console.log('📡 API response status:', res.status);
          if (res.status === 304) {
            console.log('📡 304 Not Modified - trying to parse anyway');
            return res.text().then(text => text ? JSON.parse(text) : {});
          }
          return res.json();
        })
        .then(data => {
          console.log('📥 Raw data received:', data);
          const messages = data.messages || data || [];
          console.log('📥 Messages to process:', messages);
          const existingMessages = messages.map(msg => ({
            ...msg,
            type: 'group_message',
            group_id: parseInt(groupId),
            timestamp: msg.created_at || msg.timestamp
          }));
          console.log('Mapped existing messages:', existingMessages);
          // Sort by timestamp (oldest first) for consistent chat order
          const sortedMessages = existingMessages.sort((a, b) => 
            new Date(a.timestamp) - new Date(b.timestamp)
          );
          console.log('Setting sorted messages:', sortedMessages);
          setGroupMessages(sortedMessages);
        })
        .catch(error => {
          console.error('Error loading messages:', error);
          setGroupMessages([]);
        });
    }
  }, [isOpen, groupId]);  // Add new real-time messages to existing ones
  useEffect(() => {
    if (groupId && messages) {
      const newMessages = messages.filter(msg =>
        msg.type === 'group_message' && msg.group_id === parseInt(groupId)
      );
      
      if (newMessages.length > 0) {
        setGroupMessages(prev => {
          // Merge and dedupe by from+content and timestamp proximity (3s window)
          const combined = [...prev, ...newMessages];
          const unique = [];
          for (const m of combined) {
            try {
              const isSeen = unique.some(u => {
                if (!u) return false;
                if ((u.from || u.From) !== (m.from || m.From)) return false;
                if ((u.content || u.Content) !== (m.content || m.Content)) return false;
                const t1 = new Date(u.timestamp || u.Timestamp || Date.now()).getTime();
                const t2 = new Date(m.timestamp || m.Timestamp || Date.now()).getTime();
                return Math.abs(t1 - t2) < 3000; // within 3 seconds => duplicate
              });
              if (!isSeen) unique.push(m);
            } catch (err) {
              unique.push(m);
            }
          }
          return unique.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
        });
      }
    }
  }, [messages, groupId]);

  useEffect(() => {
    if (chatWindowRef.current) {
      chatWindowRef.current.scrollTop = chatWindowRef.current.scrollHeight;
    }
  }, [groupMessages]);

  // Close emoji picker when clicking outside
  useEffect(() => {
    const handler = (e) => {
      if (!showEmojiPicker) return;
      const picker = document.getElementById('group-emoji-picker');
      const button = document.getElementById('group-emoji-button');
      if (picker && !picker.contains(e.target) && button && !button.contains(e.target)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('click', handler);
    return () => document.removeEventListener('click', handler);
  }, [showEmojiPicker]);

  // Dynamically load emoji-picker-react on first open
  useEffect(() => {
    let mounted = true;
    if (showEmojiPicker && !EmojiPicker) {
      (async () => {
        try {
          const mod = await import('emoji-picker-react');
          if (mounted) setEmojiPicker(() => mod.default || mod);
        } catch (err) {
          console.error('Failed to load emoji picker-react:', err);
        }
      })();
    }
    return () => { mounted = false; };
  }, [showEmojiPicker, EmojiPicker]);

  const handleSendMessage = (e) => {
    e.preventDefault();
    if (!messageInput.trim() || !connected) return;
    
    const messageData = {
      type: 'group_message',
      group_id: parseInt(groupId),
      from: userId,
      username: user?.username || `User ${userId}`,
      content: messageInput.trim(),
      timestamp: new Date().toISOString()
    };
    
    if (sendMessage(messageData)) {
      // optimistic append so emoji appears immediately
      setGroupMessages(prev => {
        const updated = [...prev, messageData];
        return updated.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
      });
      setMessageInput('');
    }
  };

  const insertAtCaret = (text) => {
    const input = inputRef.current;
    if (!input) {
      setMessageInput(prev => prev + text);
      return;
    }
    const start = input.selectionStart ?? input.value.length;
    const end = input.selectionEnd ?? input.value.length;
    const newVal = input.value.slice(0, start) + text + input.value.slice(end);
    setMessageInput(newVal);
    setTimeout(() => {
      const pos = start + text.length;
      input.focus();
      try { input.setSelectionRange(pos, pos); } catch (err) {}
    }, 0);
  };

  if (!isOpen) return null;

  return (
    <div className="group-chat-modern">
      {/* Modern Chat Header */}
      <div className="group-chat-header-modern">
        <div className="group-chat-header-content">
          <div className="group-chat-header-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
          <div className="group-chat-header-info">
            <h3 className="group-chat-header-title">Group Chat</h3>
            <div className="group-chat-status">
              <span className={`group-chat-status-dot ${connected ? 'connected' : 'disconnected'}`}></span>
              <span className="group-chat-status-text">
                {connected ? 'Connected' : 'Connecting...'}
              </span>
            </div>
          </div>
        </div>
        <button 
          onClick={onClose} 
          className="group-chat-close-btn"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>
      
      {/* Modern Chat Messages */}
      <div 
        ref={chatWindowRef} 
        className="group-chat-messages-modern"
      >
        {groupMessages.length === 0 ? (
          <div className="group-chat-empty-state">
            <div className="group-chat-empty-icon">
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M9 9h6M9 13h3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <h4>No messages yet</h4>
            <p>Start the conversation!</p>
          </div>
        ) : (
          <div className="group-chat-messages-list">
            {groupMessages.map((msg, index) => {
              const isMyMessage = msg.from === userId;
              const showUsername = !isMyMessage && (index === 0 || groupMessages[index - 1].from !== msg.from);
              
              return (
                <div 
                  key={index} 
                  className={`group-chat-message-modern ${isMyMessage ? 'my-message' : 'other-message'}`}
                >
                  {showUsername && !isMyMessage && (
                    <div className="message-username">{msg.username || `User ${msg.from}`}</div>
                  )}
                  <div className="message-bubble-modern">
                    <div className="message-content">{msg.content}</div>
                    <div className="message-time">
                      {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

        {/* Modern Chat Input */}
        <form 
          onSubmit={handleSendMessage} 
          className="group-chat-input-modern"
        >
          <div className="chat-input-wrapper">
            <button
              id="group-emoji-button"
              type="button"
              onClick={() => setShowEmojiPicker(s => !s)}
              title="Insert emoji"
              disabled={!connected}
              className="chat-emoji-btn"
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <path d="M8 14s1.5 2 4 2 4-2 4-2M9 9h.01M15 9h.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
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
            
            <button 
              type="submit"
              disabled={!connected || !messageInput.trim()}
              className={`chat-send-btn ${connected && messageInput.trim() ? 'active' : ''}`}
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </button>
          </div>

          {showEmojiPicker && (
            <div id="group-emoji-picker" className="chat-emoji-picker-modern">
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
                <div className="emoji-loading">Loading emojis…</div>
              )}
            </div>
          )}
        </form>
      </div>
    );
  }
  
  