import Head from 'next/head';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export default function MainPage() {
  const router = useRouter();
  const { user, loading, logout } = useSession();
  const { onlineUsers } = useWebSocketContext();

  // Ensure body-level background is applied (helps when child elements cover container)
  useEffect(() => {
    document.body.classList.add('index-page-bg');
    return () => document.body.classList.remove('index-page-bg');
  }, []);

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!loading && !user) {
      console.log('No user found, redirecting to login...');
      setTimeout(() => {
        router.push('/login');
      }, 100);
    } else if (user) {
      console.log('User found:', user);
    }
  }, [user, loading, router]);

  const handleLogout = async () => {
    const success = await logout();
    if (success) {
      router.push('/login');
    }
  };

  // Show loading while checking authentication
  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-text">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (!user) {
    return null;
  }

  return (
  <div className="layout-container index-page-bg">
      <Head>
        <title>Social Network - Dashboard</title>
      </Head>
      
      {/* Simple Header for Home Page Only */}
      <nav className="navbar">
        <div className="nav-container">
          <div className="nav-logo">
            <span style={{ cursor: 'default' }}>SocialNet</span>
          </div>
          
          <div className="nav-menu">
            <div className="nav-user-menu">
              <span className="nav-welcome">Welcome, {user.username}</span>
              <button
                onClick={handleLogout}
                className="logout-btn"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="main-content">
        <div className="dashboard-container">
          <div className="dashboard-content">
            
            {/* Professional Hero Section */}
            <div className="dashboard-hero">
              {/* Floating Particles Background */}
              <div className="floating-particles" />
              
              {/* Premium Logo */}
              <div className="inline-flex items-center justify-center w-36 h-36 bg-white/90 backdrop-blur-lg rounded-full shadow-2xl mb-10 animate-float animate-pulse-glow border-4 border-white/30">
                <div 
                  className="w-24 h-24 rounded-full animate-gentle-rotate flex items-center justify-center"
                  style={{ 
                    background: 'var(--gradient-cosmic)',
                    boxShadow: '0 0 40px rgba(102, 126, 234, 0.5)'
                  }}
                >
                  <div className="w-12 h-12 bg-white rounded-xl flex items-center justify-center">
                    {/* <div className="text-2xl font-bold text-blue-600">S</div> */}
                  </div>
                </div>
              </div>
              
              {/* Professional Welcome */}
              <div className="relative max-w-5xl mx-auto">
                <h1 className="dashboard-title">
                  Welcome back, {user?.username}!
                </h1>
                
                {/* Floating Decorative Elements */}
                <div className="absolute -top-8 -right-12 w-12 h-12 bg-gradient-to-r from-pink-400 to-red-500 rounded-xl opacity-60 animate-bounce"></div>
                <div className="absolute -top-4 -left-16 w-10 h-10 bg-gradient-to-r from-blue-400 to-purple-500 rounded-full opacity-50 animate-pulse"></div>
                <div className="absolute -bottom-6 right-20 w-8 h-8 bg-gradient-to-r from-yellow-400 to-orange-500 rounded-full opacity-45 animate-float"></div>
                
                {/* <p className="dashboard-subtitle mt-8">
                  Your <span className="font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">premium social hub</span> awaits! 
                  Connect with amazing people, share incredible stories, and discover 
                  <span className="font-bold bg-gradient-to-r from-pink-600 to-purple-600 bg-clip-text text-transparent"> magical content</span> 
                  from your vibrant community!
                </p> */}
              </div>
              
              {/* Creative Stories Stats */}
              <div className="bg-white/90 backdrop-blur-2xl rounded-3xl p-8 shadow-2xl border border-white/30 relative overflow-hidden group hover:scale-105 transition-all duration-500 animate-float" style={{animationDelay: '1s'}}>
                {/* Background Decoration */}
                <div className="absolute top-0 right-0 w-32 h-32 rounded-full opacity-15 animate-pulse" style={{ background: 'var(--gradient-nature)' }}></div>
                <div className="dashboard-card-decoration-small"></div>
                
                <div className="flex flex-col items-center text-center relative z-10">
                  <div className="w-20 h-20 rounded-2xl flex items-center justify-center mb-6 animate-gentle-rotate group-hover:scale-110 transition-transform shadow-lg" style={{ background: 'var(--gradient-nature)' }}>
                   
                  </div>
                  
                  <div className="text-5xl font-black mb-2 bg-gradient-to-r from-emerald-600 via-teal-600 to-cyan-700 bg-clip-text text-transparent">
                    ∞
                  </div>
                  
                  <h3 className="text-xl font-bold text-gray-800 mb-3">Stories to Share</h3>
                  
                  <div className="flex items-center gap-3 bg-emerald-50/80 px-4 py-2 rounded-full backdrop-blur-sm">
                    <div className="w-4 h-4 bg-emerald-600 rounded-full animate-bounce"></div>
                    <span className="text-sm text-emerald-700 font-semibold">Unlimited creativity!</span>
                  </div>
                </div>
              </div>
              
              {/* Magic Experiences Stats */}
              <div className="bg-white/90 backdrop-blur-2xl rounded-3xl p-8 shadow-2xl border border-white/30 relative overflow-hidden group hover:scale-105 transition-all duration-500 animate-float" style={{animationDelay: '2s'}}>
                {/* Background Decoration */}
                <div className="absolute top-0 right-0 w-32 h-32 rounded-full opacity-15 animate-pulse" style={{ background: 'var(--gradient-magic)' }}></div>
                <div className="dashboard-card-decoration-small"></div>
                
                <div className="flex flex-col items-center text-center relative z-10">
                  <div className="w-20 h-20 rounded-2xl flex items-center justify-center mb-6 animate-gentle-rotate group-hover:scale-110 transition-transform shadow-lg" style={{ background: 'var(--gradient-magic)' }}>
                  </div>
                  
                  <div className="text-4xl font-black mb-2 bg-gradient-to-r from-purple-600 via-pink-600 to-purple-800 bg-clip-text text-transparent">
                    Magic
                  </div>
                  
                  <h3 className="text-xl font-bold text-gray-800 mb-3">New Experiences</h3>
                  
                  <div className="flex items-center gap-3 bg-purple-50/80 px-4 py-2 rounded-full backdrop-blur-sm">
                    <div className="w-4 h-4 bg-purple-600 rounded-full animate-pulse"></div>
                    <span className="text-sm text-purple-700 font-semibold">Every moment!</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Professional Action Cards Section */}
            <div className="dashboard-actions-grid">
              {/* Explore Posts - Ocean Theme */}
            <button 
              onClick={() => router.push('/posts')}
                className="dashboard-card dashboard-card-blue group"
              >
                <div className="dashboard-card-decoration"></div>
                <div className="dashboard-card-decoration-small"></div>
                
                <div className="dashboard-card-icon animate-float group-hover:scale-125 transition-all duration-500">
                  <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center text-2xl font-black text-white">
                    POST
                  </div>
                </div>
                
                <h3 className="dashboard-card-title">Explore Posts</h3>
                <p className="dashboard-card-description">
                  Discover <span className="font-bold text-white/90">incredible content</span> from your amazing network! 
                  Dive into stories that inspire!
                </p>
                
                <div className="dashboard-card-action group-hover:scale-110">
                  Let's explore
                </div>
            </button>
            
              {/* Create Magic - Nature Theme */}
            <button 
              onClick={() => router.push('/createPost')}
                className="dashboard-card dashboard-card-green group"
              >
                <div className="dashboard-card-decoration"></div>
                <div className="dashboard-card-decoration-small"></div>
                
                <div className="dashboard-card-icon animate-float group-hover:scale-125 transition-all duration-500">
                  <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center text-lg font-black text-white">
                    CREATE
                  </div>
                </div>
                
                <h3 className="dashboard-card-title">Create Content</h3>
                <p className="dashboard-card-description">
                  Share your <span className="font-bold text-white/90">brilliant ideas</span> with the world! 
                  Inspire others with your creativity!
                </p>
                
                <div className="dashboard-card-action group-hover:scale-110">
                  Start creating
                </div>
            </button>
            
              {/* Join Communities - Magic Theme */}
            <button 
              onClick={() => router.push('/groups')}
                className="dashboard-card dashboard-card-purple group"
              >
                <div className="dashboard-card-decoration"></div>
                <div className="dashboard-card-decoration-small"></div>
                
                <div className="dashboard-card-icon animate-float group-hover:scale-125 transition-all duration-500">
                  <div className="w-16 h-16 bg-white/20 rounded-2xl flex items-center justify-center text-lg font-black text-white">
                    GROUP
                  </div>
                </div>
                
                <h3 className="dashboard-card-title">Join Communities</h3>
                <p className="dashboard-card-description">
                  Find your <span className="font-bold text-white/90">perfect tribe</span> and connect with 
                  incredible people! Build lasting bonds!
                </p>
                
                <div className="dashboard-card-action group-hover:scale-110">
                  Explore groups
                </div>
            </button>
          </div>

            
          {/* Online Users Section */}
          <div className="online-users-section">
            <h2>
              Community Members Online ({onlineUsers.filter(u => u.id !== user?.userID).length})
            </h2>
            {onlineUsers.filter(u => u.id !== user?.userID).length === 0 ? (
              <div className="text-center py-12">
                <div className="w-16 h-16 bg-gray-200 rounded-full mx-auto mb-4 flex items-center justify-center text-2xl font-bold text-gray-500">
                  0
                </div>
                <p className="text-lg font-medium text-gray-600">No users online right now</p>
                <p className="text-gray-500 mt-2">Be the first to start a conversation!</p>
              </div>
            ) : (
              <div className="online-users-list">
                {onlineUsers
                  .filter(u => u.id !== user?.userID)
                  .map((userData) => (
                    <div 
                      key={userData.id} 
                      className="online-user-item group cursor-pointer"
                      onClick={() => router.push('/users')}
                    >
                      <div className="user-info">
                        <div className="user-avatar-small">
                          <span>
                            {userData.username?.charAt(0)?.toUpperCase() || 'U'}
                          </span>
                        </div>
                        <div className="flex-1">
                        <span className="username">{userData.username}</span>
                          <div className="flex items-center gap-2 mt-1">
                            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                            <span className="text-xs text-gray-500 font-medium">Active now</span>
                          </div>
                        </div>
                      </div>
                      <div className="w-8 h-8 bg-green-100 text-green-600 rounded-full flex items-center justify-center font-bold group-hover:scale-110 transition-transform text-xs">
                        ONLINE
                      </div>
                    </div>
                  ))}
              </div>
            )}
          </div>

            {/* Premium Call-to-Action Footer */}
            <div className="text-center mt-20 pb-12">
              <div className="bg-white/90 backdrop-blur-2xl rounded-3xl p-10 shadow-2xl border border-white/30 relative overflow-hidden max-w-2xl mx-auto">
                {/* Background Effects */}
                <div className="absolute inset-0 bg-gradient-to-r from-blue-50/50 via-purple-50/50 to-pink-50/50 rounded-3xl"></div>
                <div className="absolute top-2 right-2 w-20 h-20 rounded-full opacity-10 animate-pulse" style={{ background: 'var(--gradient-cosmic)' }}></div>
                
                <div className="relative z-10">
                  {/* <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-500 rounded-2xl mx-auto mb-6 flex items-center justify-center text-white font-black text-lg animate-bounce">
                    GO
                  </div> */}
                  <h3 className="text-2xl font-bold text-gray-800 mb-4">Ready for your next adventure?</h3>
                  <p className="text-lg text-gray-600 font-medium mb-8">
                    Your journey to <span className="font-bold text-gradient">amazing connections</span> starts now!
                  </p>
                  
                  <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                    <button 
                      onClick={() => router.push('/users')}
                      className="btn btn-primary btn-lg group relative overflow-hidden"
                    >
                      Find Amazing Friends
                    </button>
                    
                    {/* <button 
                      onClick={() => router.push('/createPost')}
                      className="btn btn-success btn-lg group relative overflow-hidden"
                    >
                      Share Your Story
                    </button> */}
                  </div>
          </div>
        </div>
        </div>
            
          </div> {/* End dashboard-content */}
        </div> {/* End dashboard-container */}
      </main>
    </div>
  );
}