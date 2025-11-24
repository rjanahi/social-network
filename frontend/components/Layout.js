import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import Notifications from './Notifications';

const Layout = ({ children }) => {
  const { user, loading, logout } = useSession();
  const { onlineUsers, connected } = useWebSocketContext();
  const [showNotifications, setShowNotifications] = useState(false);
  const router = useRouter();

  const handleLogout = async () => {
    const success = await logout();
    if (success) {
      router.push('/login');
    }
  };

  const navigateTo = (path) => {
    router.push(path);
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-text">Loading...</div>
      </div>
    );
  }

  return (
    <div className="layout-container">
      {/* Navigation Bar */}
      <nav className="navbar">
        <div className="nav-container">
          <div className="nav-logo">
            <Link href="/">
              SocialNet
            </Link>
          </div>
          
          <div className="nav-menu">
            {!user ? (
              <>
                <button 
                  onClick={() => navigateTo('/login')} 
                  className="nav-link"
                >
                  Log In
                </button>
                <button 
                  onClick={() => navigateTo('/signUp')} 
                  className="nav-link"
                >
                  Sign Up
                </button>
              </>
            ) : (
              <>
                <button 
                  onClick={() => navigateTo('/posts')} 
                  className="nav-link"
                >
                  Posts
                </button>
                {/* <button 
                  onClick={() => navigateTo('/createPost')} 
                  className="nav-link"
                >
                  Create Post
                </button> */}
                <button 
                  onClick={() => navigateTo('/groups')} 
                  className="nav-link"
                >
                  Groups
                </button>
                <button 
                  onClick={() => navigateTo('/users')} 
                  className="nav-link"
                >
                  Find Users
                </button>
                <button 
                  onClick={() => navigateTo('/myProfile')} 
                  className="nav-link"
                >
                  Profile
                </button>
                
                {/* User Menu */}
                <div className="nav-user-menu">
                  <Notifications />
                  <span className="nav-welcome">Welcome, {user.username}</span>
                  <button
                    onClick={handleLogout}
                    className="logout-btn"
                  >
                    Logout
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="main-content">
        {children}
      </main>
    </div>
  );
};

export default Layout;