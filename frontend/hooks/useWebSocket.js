import { useWebSocketContext } from '../contexts/WebSocketContext';

// This hook now simply exports the WebSocket context
// All pages/components will share the same WebSocket connection
export const useWebSocket = (userID) => {
  return useWebSocketContext();
};
