import { useState, useEffect, useCallback } from 'react';

export const useSession = () => {
  const [session, setSession] = useState({
    isLoggedIn: false,
    userID: null,
    username: null,
    loading: true
  });

  const checkSession = useCallback(async () => {
    try {
      setSession(prev => ({ ...prev, loading: true }));
      
      const response = await fetch('/api/session', {
        method: 'GET',
        credentials: 'include'
      });
      
      const data = await response.json();
      console.log("SESSION CHECK RESPONSE:", data);
      
      if (data.loggedIn && typeof data.userID !== "undefined") {
        setSession({
          isLoggedIn: true,
          userID: data.userID,
          username: data.username,
          loading: false
        });
        
        // Store in localStorage for persistence
        if (typeof window !== 'undefined') {
          localStorage.setItem('username', data.username);
          window.userID = data.userID;
        }
      } else {
        setSession({
          isLoggedIn: false,
          userID: null,
          username: null,
          loading: false
        });
      }
    } catch (error) {
      console.error('Session check failed:', error);
      setSession({
        isLoggedIn: false,
        userID: null,
        username: null,
        loading: false
      });
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include'
      });
      
      const data = await response.json();
      console.log(data.message);
      
      setSession({
        isLoggedIn: false,
        userID: null,
        username: null,
        loading: false
      });
      
      // Clear localStorage
      if (typeof window !== 'undefined') {
        localStorage.removeItem('username');
        window.userID = null;
      }
      
      return true;
    } catch (error) {
      console.error('Logout failed:', error);
      return false;
    }
  }, []);

  // Check session on mount
  useEffect(() => {
    checkSession();
  }, [checkSession]);

  return {
    user: session.isLoggedIn ? {
      userID: session.userID,
      username: session.username
    } : null,
    loading: session.loading,
    session,
    checkSession,
    logout
  };
};