import { useState, useEffect } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Layout from '../components/Layout';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export default function Users() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const { user, loading: sessionLoading } = useSession();
  const { subscribe } = useWebSocketContext();
  const router = useRouter();

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/users', {
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        console.log('[Users] Fetched users:', data.users);
        setUsers(data.users || []);
      } else {
        console.error('Users API failed with status:', response.status);
        const errorText = await response.text();
        console.error('Error response:', errorText);
      }
    } catch (error) {
      console.error('Failed to fetch users:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!sessionLoading && user?.userID) {
      fetchUsers();
    }
  }, [sessionLoading, user?.userID]);

  // Listen for real-time follow/unfollow updates
  useEffect(() => {
    if (!user) return;

    const unsubscribe = subscribe((message) => {
      console.log('[Users] Received WebSocket message:', message);

      // Handle follow status updates and follow request responses
      if (message.type === 'follow_update' || message.type === 'new_follower' || message.type === 'unfollow_update' || message.type === 'follow_request_response' || message.type === 'follow_request' || message.type === 'privacy_update') {
        // Refresh the users list to get updated follow/request status
        fetchUsers();
      }
    });

    return () => unsubscribe();
  }, [user, subscribe]);



  const handleFollow = async (username) => {
    try {
      const response = await fetch('/api/follow-user', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ username })
      });

      if (response.ok) {
        const data = await response.json();
        
        // Show appropriate message based on response
        if (data.status === 'pending') {
          console.log('Follow request sent:', data.message);
        } else if (data.status === 'following') {
          console.log('Now following:', data.message);
        }
        
        // Refresh the list to show updated status
        fetchUsers();
      } else {
        const errorText = await response.text();
        console.error('Failed to follow user:', response.status, errorText);
        
        // If already following, just refresh
        if (errorText.includes('Already following')) {
          fetchUsers();
        }
      }
    } catch (error) {
      console.error('Error following user:', error);
    }
  };

  const handleUnfollow = async (username) => {
    try {
      const response = await fetch('/api/unfollow-user', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ username })
      });

      if (response.ok) {
        const data = await response.json();
        console.log('Unfollow response:', data.message || data);
        fetchUsers(); // Refresh the list
      } else {
        console.error('Failed to unfollow user:', response.status);
      }
    } catch (error) {
      console.error('Error unfollowing user:', error);
    }
  };



  if (sessionLoading) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Loading...</div>
        </div>
      </Layout>
    );
  }

  if (!user) {
    if (typeof window !== 'undefined') {
      router.push('/login');
    }
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Redirecting to login...</div>
        </div>
      </Layout>
    );
  }

  // Filter users by search
  const filteredUsers = users.filter(u => {
    const term = search.trim().toLowerCase();
    if (!term) return true;
    return (
      u.username?.toLowerCase().includes(term) ||
      u.firstname?.toLowerCase().includes(term) ||
      u.lastname?.toLowerCase().includes(term)
    );
  });

  return (
    <Layout>
      <Head>
        <title>Find Amazing People - Connect & Inspire</title>
        <meta name="description" content="Discover incredible people and build lasting friendships!" />
      </Head>
      
      <div className="users-page-bg">
        <div className="container-professional">
          {/* Professional Users Header */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center justify-center w-28 h-28 bg-white/90 backdrop-blur-lg rounded-3xl shadow-xl mb-8 animate-float animate-pulse-glow">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600"></div>
            </div>
            
            <h1 className="text-5xl font-bold text-gradient mb-6">
              Discover Amazing People
            </h1>
            
            <p className="text-xl text-gray-700 font-medium mb-8 max-w-3xl mx-auto">
              Connect with <span className="font-bold text-gradient">incredible souls</span> who share your passions! 
              Build friendships that create <span className="font-bold text-gradient">magical memories</span>!
            </p>
          </div>
          {/* Professional Search Section */}
          <div className="bg-white/90 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/30 p-8 mb-12 relative overflow-hidden">
            <div className="absolute top-0 left-0 right-0 h-1" style={{ background: 'var(--gradient-cosmic)' }}></div>
            
            <div className="text-center mb-6">
              <p className="text-gray-600 font-medium">Search for amazing people by name or username</p>
            </div>
            
            <div className="max-w-lg mx-auto relative">
              <input
                type="text"
                placeholder="Search for incredible people..."
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="w-full px-8 py-4 border-2 border-gray-200 rounded-2xl bg-gray-50/50 font-semibold text-lg transition-all duration-300 focus:border-purple-500 focus:bg-white focus:shadow-xl focus:transform focus:scale-105 outline-none backdrop-blur-sm"
              />
            </div>
          </div>
          {/* Professional Users List */}
          <div className="space-y-6">
            {loading ? (
              <div className="bg-white rounded-2xl p-12 shadow-lg border border-gray-200 text-center">
                <div className="text-6xl mb-6">
                  <div className="w-16 h-16 mx-auto bg-gradient-to-r from-blue-500 to-purple-600 rounded-full animate-spin"></div>
                </div>
                <h3 className="text-2xl font-bold text-gray-700 mb-4">Finding amazing people...</h3>
                <p className="text-lg text-gray-600">Discovering connections for you</p>
              </div>
            ) : (
              <div className="space-y-6">
                {filteredUsers.length === 0 ? (
                  <div className="bg-white rounded-2xl p-16 shadow-lg border border-gray-200 text-center">
                    <div className="text-8xl mb-8">
                      <div className="w-24 h-24 mx-auto bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-4xl font-bold">?</div>
                    </div>
                    <h3 className="text-3xl font-bold text-gray-800 mb-4">No matches found</h3>
                    <p className="text-xl text-gray-600 mb-8">
                      Try a different search or explore all our <span className="font-bold text-blue-600">amazing members</span>
                    </p>
                    <button 
                      onClick={() => setSearch('')}
                      className="btn btn-primary btn-lg"
                    >
                      Show All Users
                    </button>
                  </div>
                ) : (
                  <div className="grid gap-6">
                    {filteredUsers.map((u) => (
                      <div
                        key={u.id}
                        className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8 hover:shadow-xl hover:border-blue-400 transition-all duration-300 cursor-pointer"
                        onClick={() => router.push(`/profile/${encodeURIComponent(u.username)}`)}
                      >
                        {/* User Card Content */}
                        <div className="flex items-center justify-between">
                          {/* User Info Section */}
                          <div className="flex items-center gap-6">
                            {/* Professional Avatar */}
                            <div className="relative">
                              <img
                                src={u.avatar || '/img/avatars/images.png'}
                                alt={u.username}
                                className="rounded-full object-cover border-2 border-white shadow-sm transition duration-200"
                                style={{ width: 64, height: 64, minWidth: 64, minHeight: 64 }}
                                onError={e => { e.target.src = '/img/avatars/images.png'; }}
                              />
                              {/* Online Status Indicator */}
                              <div className="absolute -top-1 -right-1 w-6 h-6 bg-green-500 rounded-full border-2 border-white animate-pulse shadow-md"></div>
                            </div>
                            
                            {/* User Details */}
                            <div className="flex-1">
                              <h3 className="text-2xl font-bold text-gray-900 mb-2 hover:text-blue-600 transition-colors">
                                {u.firstname} {u.lastname}
                              </h3>
                              <p className="text-lg text-gray-600 font-medium mb-3">@{u.username}</p>
                              
                              {/* Status Badges */}
                              <div className="flex flex-wrap items-center gap-3">
                                {u.isPrivate && (
                                  <span className="inline-flex items-center gap-2 px-4 py-2 bg-red-50 text-red-700 rounded-full border border-red-200 font-semibold">
                                    <span>Private</span>
                                  </span>
                                )}
                                {u.isFollowing && (
                                  <span className="inline-flex items-center gap-2 px-4 py-2 bg-emerald-50 text-emerald-700 rounded-full border border-emerald-200 font-semibold animate-pulse">
                                    <span>Following</span>
                                  </span>
                                )}
                                {u.requestStatus === 'pending' && (
                                  <span className="inline-flex items-center gap-2 px-4 py-2 bg-yellow-50 text-yellow-700 rounded-full border border-yellow-200 font-semibold">
                                    <span>Request Sent</span>
                                  </span>
                                )}
                              </div>
                            </div>
                          </div>
                          
                          {/* Action Button Section */}
                          <div className="flex items-center gap-3" onClick={e => e.stopPropagation()}>
                            {u.isFollowing ? (
                              <button
                                onClick={() => handleUnfollow(u.username)}
                                className="btn btn-danger btn-lg group relative overflow-hidden"
                              >
                                Unfollow
                              </button>
                            ) : u.requestStatus === 'pending' ? (
                              <button
                                disabled
                                className="btn btn-warning btn-lg opacity-60 cursor-not-allowed"
                              >
                                Request Sent
                              </button>
                            ) : (
                              <button
                                onClick={() => handleFollow(u.username)}
                                className="btn btn-success btn-lg group relative overflow-hidden"
                              >
                                {u.isPrivate ? 'Request' : 'Follow'}
                              </button>
                            )}
                          </div>
                        </div>
                        
                        {/* Hover Effect Decorations */}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
      
    {/* Modal removed — users navigate to /profile/[username] page now */}
      </Layout>
    );
  }