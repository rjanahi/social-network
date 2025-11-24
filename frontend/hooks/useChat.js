import { useState } from 'react';

export const useChat = () => {
  const [chatOpen, setChatOpen] = useState(false);
  const [chatWith, setChatWith] = useState(null);

  const startChat = (user) => {
    setChatWith(user);
    setChatOpen(true);
  };

  const closeChat = () => {
    setChatOpen(false);
    setChatWith(null);
  };

  return {
    chatOpen,
    chatWith,
    startChat,
    closeChat,
    setChatOpen,
    setChatWith
  };
};