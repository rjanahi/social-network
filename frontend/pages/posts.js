import { useState, useEffect } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Layout from '../components/Layout';
import Chat from '../components/Chat';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import { useChat } from '../hooks/useChat';

export default function Posts() {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [followingUsers, setFollowingUsers] = useState([]);
  const { user, loading: sessionLoading } = useSession();
  const { onlineUsers, subscribe } = useWebSocketContext();
  const { chatOpen, chatWith, startChat, setChatOpen } = useChat();
  const router = useRouter();
  const [mobileChatOpen, setMobileChatOpen] = useState(false);

  const fetchPosts = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/posts', {
        method: 'GET',
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        // Handle null or undefined data gracefully
        setPosts(data?.posts || []);
      }
    } catch (error) {
      console.error('Failed to fetch posts:', error);
      setPosts([]); // Set empty array on error
    } finally {
      setLoading(false);
    }
  };

  const fetchFollowingUsers = async () => {
    try {
      const response = await fetch('/api/user-followers', {
        method: 'GET',
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        setFollowingUsers(data?.following || []);
      }
    } catch (error) {
      console.error('Failed to fetch following users:', error);
      setFollowingUsers([]);
    }
  };

  // Listen for real-time post updates
  useEffect(() => {
    if (!user) return;

    const unsubscribe = subscribe((message) => {
      console.log('[Posts] Received WebSocket message:', message);


      // Handle normal post creation
      if (message.type === 'new_post') {
        try {
          const postData = JSON.parse(message.content);
          // If the post is from the current user and avatar_url is missing, inject user's avatar
          let postWithAvatar = postData;
          if (user && message.from === user.userID && (!postData.avatar_url || postData.avatar_url === '' || postData.avatar_url === '/img/avatars/images.png')) {
            postWithAvatar = { ...postData, avatar_url: user.avatar };
          }
          setPosts(prev => {
            // If post already exists, update avatar_url if needed (for all users)
            if (prev.some(p => p.id === postWithAvatar.id)) {
              return prev.map(p =>
                p.id === postWithAvatar.id && (!p.avatar_url || p.avatar_url === '' || p.avatar_url === '/img/avatars/images.png')
                  ? { ...p, avatar_url: postWithAvatar.avatar_url }
                  : p
              );
            }
            return [postWithAvatar, ...prev];
          });
          if (message.username && message.from !== user.userID) {
            console.log(`New post from ${message.username}: ${postData.title}`);
          }
        } catch (e) {
          console.error('[Posts] Failed to parse post data:', e);
        }
      }

      // Handle group post creation in real-time
      else if (message.type === 'new_groupPost') {
        try {
          const groupPostData = JSON.parse(message.content);
          let groupPostWithAvatar = groupPostData;
          if (user && message.from === user.userID && (!groupPostData.avatar_url || groupPostData.avatar_url === '' || groupPostData.avatar_url === '/img/avatars/images.png')) {
            groupPostWithAvatar = { ...groupPostData, avatar_url: user.avatar };
          }
          setPosts(prev => {
            // If post already exists, update avatar_url if needed (for all users)
            if (prev.some(p => p.id === groupPostWithAvatar.id)) {
              return prev.map(p =>
                p.id === groupPostWithAvatar.id && (!p.avatar_url || p.avatar_url === '' || p.avatar_url === '/img/avatars/images.png')
                  ? { ...p, avatar_url: groupPostWithAvatar.avatar_url }
                  : p
              );
            }
            return [groupPostWithAvatar, ...prev];
          });
          if (message.username && message.from !== user.userID) {
            console.log(`New group post from ${message.username}: ${groupPostData.title}`);
          }
        } catch (e) {
          console.error('[Posts] Failed to parse group post data:', e);
        }
      }

      // Handle like updates
      else if (message.type === 'new_postLike') {
        try {
          const likeData = JSON.parse(message.content);
          setPosts(prev => prev.map(post => 
            post.id === likeData.post_id 
              ? { 
                  ...post, 
                  likes_count: likeData.likes_count,
                  dislikes_count: likeData.dislikes_count 
                }
              : post
          ));
        } catch (e) {
          console.error('[Posts] Failed to parse like data:', e);
        }
      }

      // Handle comment updates
      else if (message.type === 'new_comment' && message.post_id) {
        setPosts(prev => prev.map(post => 
          post.id === message.post_id 
            ? { ...post, comments: (post.comments || 0) + 1 }
            : post
        ));
      }

      // Handle follow-related events: refresh the following users list so
      // the chat sidebar updates immediately (no page refresh needed).
      // This includes follow, follow-request responses, follow state updates
      // and unfollow notifications so users appear/disappear in real-time.
      else if (
        message.type === 'new_follower' ||
        message.type === 'follow_request_response' ||
        message.type === 'follow_update' ||
        message.type === 'unfollow_update'
      ) {
        // Best-effort: refetch following users in background
        fetchFollowingUsers();
      }
    });

    return () => unsubscribe();
  }, [user, subscribe]);

  const handleLike = async (postId) => {
    try {
      const response = await fetch('/api/like', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          post_id: postId,
          is_like: true
        }),
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        // Update local state immediately with the response data
        setPosts(prev => prev.map(post => 
          post.id === postId 
            ? { 
                ...post, 
                likes_count: data.likes,
                dislikes_count: data.dislikes 
              }
            : post
        ));
      }
    } catch (error) {
      console.error('Failed to like post:', error);
    }
  };

  const handleDislike = async (postId) => {
    try {
      const response = await fetch('/api/like', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          post_id: postId,
          is_like: false
        }),
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        // Update local state immediately with the response data
        setPosts(prev => prev.map(post => 
          post.id === postId 
            ? { 
                ...post, 
                likes_count: data.likes,
                dislikes_count: data.dislikes 
              }
            : post
        ));
      }
    } catch (error) {
      console.error('Failed to dislike post:', error);
    }
  };

  useEffect(() => {
    fetchPosts();
    fetchFollowingUsers();
  }, []);



  if (sessionLoading) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Loading...</div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <Head>
        <title>Posts - Discover Amazing Content</title>
        <meta name="description" content="Explore incredible stories from your community!" />
      </Head>
      
      <div className="posts-page-bg">
        <div className={`posts-layout ${chatOpen ? 'chat-open' : ''}`}>
          {/* Hero Section - matching groups page style */}
          <div className="posts-hero-section">
            <div className="posts-hero-content">
              <div className="posts-hero-icon">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <polyline points="14 2 14 8 20 8" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <line x1="16" y1="13" x2="8" y2="13" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <line x1="16" y1="17" x2="8" y2="17" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <polyline points="10 9 9 9 8 9" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              
              <h1 className="posts-hero-title">
                Discover Amazing Stories
              </h1>
            

              <button 
                onClick={() => router.push('/createPost')}
                className="posts-hero-cta"
              >
                <span className="posts-hero-cta-icon">+</span>
                Create New Post
              </button>
            </div>
          </div>
          
          {/* Main Content Container */}
          <div className="posts-main-content container-professional">
            <div className="posts-container-modern">
              {loading ? (
                <div style={{
                  background: 'white',
                  borderRadius: '20px',
                  padding: '48px',
                  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.08)',
                  textAlign: 'center'
                }}>
                  <div className="text-6xl mb-6 animate-spin flex items-center justify-center" style={{animationDuration: '2s'}}>
                    <div className="w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full"></div>
                  </div>
                  <h3 className="text-2xl font-bold mb-4" style={{color: '#6b7280'}}>Loading amazing stories...</h3>
                  <p className="text-lg text-gray-600">Preparing incredible content for you!</p>
                </div>
              ) : posts.length === 0 ? (
                <div style={{
                  background: 'white',
                  borderRadius: '20px',
                  padding: '64px 32px',
                  boxShadow: '0 4px 20px rgba(0, 0, 0, 0.08)',
                  textAlign: 'center',
                  position: 'relative',
                  overflow: 'hidden'
                }}>
                  <div className="text-8xl mb-8 animate-float flex items-center justify-center">
                    <div className="w-24 h-24 bg-gradient-to-r from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center">
                      <div className="text-white text-4xl font-bold">+</div>
                    </div>
                  </div>
                  <h3 className="text-3xl font-bold mb-4" style={{color: '#4b5563'}}>No posts yet!</h3>
                  <p className="text-xl mb-8" style={{color: '#6b7280'}}>
                    Be the first to share something <span className="font-bold text-gradient">amazing</span>!
                  </p>
                  <button 
                    onClick={() => router.push('/createPost')}
                    className="btn btn-primary btn-lg"
                    style={{ 
                      background: 'linear-gradient(135deg, #3a4da5ff 0%, #927ca8ff 100%)',
                      color: 'white',
                      border: 'none',
                      padding: '14px 32px',
                      borderRadius: '50px',
                      fontSize: '16px',
                      fontWeight: '600',
                      cursor: 'pointer',
                      transition: 'all 0.3s ease',
                      boxShadow: '0 8px 24px rgba(102, 126, 234, 0.3)'
                    }}
                  >
                    Create Your First Post
                  </button>
                </div>
              ) : (
              posts.map((post) => (
                <div key={post.id} className="enhanced-post-card">
                  {/* Post Card Header */}
                  <div className="enhanced-post-header">
                    <div className="enhanced-post-author">
                      {post.avatar_url && post.avatar_url !== '/img/avatars/images.png' && post.avatar_url !== '' ? (
                        <img
                          src={post.avatar_url}
                          alt={`${post.username}'s avatar`}
                          className="enhanced-post-avatar"
                          onError={e => { e.target.src = '/img/avatars/images.png'; }}
                        />
                      ) : (
                        <div className="enhanced-post-avatar-fallback">
                          {post.firstname && post.firstname.length > 0
                            ? post.firstname[0]
                            : (post.username && post.username.length > 0
                                ? post.username[0]
                                : '?')}
                        </div>
                      )}
                      <div className="enhanced-post-author-info">
                        <h3 className="enhanced-post-username">{post.username}</h3>
                        <p className="enhanced-post-date">
                          {new Date(post.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    {post.category && (
                      <span className="enhanced-post-category">
                        {post.category}
                      </span>
                    )}
                  </div>

                  {/* Post Content */}
                  <div className="enhanced-post-body">
                    <h2 className="enhanced-post-title">{post.title}</h2>
                    <p className="enhanced-post-content">{post.content}</p>

                    {post.image && (
                      <div className="enhanced-post-image-container">
                        <img
                          src={post.image}
                          alt="Post image"
                          className="enhanced-post-image"
                        />
                      </div>
                    )}
                  </div>

                  {/* Post Actions */}
                  <div className="enhanced-post-actions">
                    <div className="enhanced-action-buttons">
                      <button
                        onClick={() => handleLike(post.id)}
                        className="enhanced-action-btn like-btn-enhanced"
                      >
                        <div className="action-icon">+</div>
                        <span>Like ({post.likes_count || 0})</span>
                      </button>
                      <button
                        onClick={() => handleDislike(post.id)}
                        className="enhanced-action-btn dislike-btn-enhanced"
                      >
                        <div className="action-icon">-</div>
                        <span>Dislike ({post.dislikes_count || 0})</span>
                      </button>
                      <button
                        onClick={() => router.push(`/comments?postId=${post.id}`)}
                        className="enhanced-action-btn comment-btn-enhanced"
                      >
                        <div className="action-icon">C</div>
                        <span>Comments ({post.comments || 0})</span>
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Right Sidebar - Chat with Following Users */}
        {user && (
            <div className="chat-sidebar">
            <h3 className="chat-sidebar-title">
              Chat ({followingUsers.length})
            </h3>
            <div className="chat-users-list">
                    {followingUsers.length > 0 ? (
                followingUsers.map((followingUser) => {
                  const isOnline = onlineUsers.some(u => u.id === followingUser.id);
                  return (
                    <div
                      key={followingUser.id}
                      className="chat-user-item"
                      role="button"
                      tabIndex={0}
                      onClick={() => startChat(followingUser)}
                      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') startChat(followingUser); }}
                    >
                      <div className="chat-user-info">
                        <div className={`chat-status-dot ${isOnline ? 'online' : 'offline'}`}></div>
                        <div className="user-avatar-small">
                          <span>{followingUser.emoji || (followingUser.username ? followingUser.username.charAt(0).toUpperCase() : '?')}</span>
                        </div>
                        <span className="chat-username">{followingUser.username}</span>
                      </div>
                    </div>
                  );
                })
              ) : (
                <div className="chat-empty-state">
                  <p>No users to chat with</p>
                  <p className="chat-empty-subtext">Follow users to start chatting</p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Mobile chat drawer toggle (visible only on small screens via CSS) */}
        {user && (
          <>
            <button
              className="mobile-chat-button"
              onClick={() => setMobileChatOpen(true)}
              aria-label="Open chat"
            >
              💬
            </button>

            <div className={`mobile-chat-drawer ${mobileChatOpen ? 'open' : ''}`} role="dialog" aria-modal="true">
              <div className="mobile-chat-drawer-header">
                <h4>Chat</h4>
                <button className="mobile-chat-close" onClick={() => setMobileChatOpen(false)}>×</button>
              </div>
              <div className="mobile-chat-users">
                {followingUsers.length > 0 ? (
                  followingUsers.map((followingUser) => (
                    <div
                      key={followingUser.id}
                      className="chat-user-item"
                      role="button"
                      tabIndex={0}
                      onClick={() => { startChat(followingUser); setMobileChatOpen(false); }}
                      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { startChat(followingUser); setMobileChatOpen(false); } }}
                    >
                      <div className="chat-user-info">
                        <div className={`chat-status-dot ${onlineUsers.some(u => u.id === followingUser.id) ? 'online' : 'offline'}`}></div>
                        <div className="user-avatar-small"><span>{followingUser.emoji || (followingUser.username ? followingUser.username.charAt(0).toUpperCase() : '?')}</span></div>
                        <span className="chat-username">{followingUser.username}</span>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="chat-empty-state">
                    <p>No users to chat with</p>
                    <p className="chat-empty-subtext">Follow users to start chatting</p>
                  </div>
                )}
              </div>
            </div>
          </>
        )}
      </div>
      </div> {/* Close posts-page-bg */}

      {/* Chat Component */}
      <Chat
        isOpen={chatOpen}
        onClose={() => setChatOpen(false)}
        chatWith={chatWith}
        userID={user?.userID}
      />
    </Layout>
  );
}
