import { useState, useEffect } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Layout from '../components/Layout';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import { useChat } from '../hooks/useChat';
import Chat from '../components/Chat';

export default function MyProfile() {
  const [profile, setProfile] = useState(null);
  const [posts, setPosts] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [pendingRequests, setPendingRequests] = useState([]);
  const [isPrivate, setIsPrivate] = useState(false);
  const [loading, setLoading] = useState(true);
  const [editMode, setEditMode] = useState(false);
  const [editForm, setEditForm] = useState({
    firstname: '',
    lastname: '',
    nickname: '',
    bio: '',
    email: '',
    age: '',
    gender: '',
    dateOfBirth: '',
    avatarFile: null
  });

 const [avatarRefresh, setAvatarRefresh] = useState(Date.now());
  const { user, loading: sessionLoading } = useSession();
  const { onlineUsers, subscribe } = useWebSocketContext();
  const { chatOpen, chatWith, startChat, setChatOpen } = useChat();
  const router = useRouter();

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/profile', {
        method: 'GET',
        credentials: 'include'
      });

      if (response.ok) {
        const data = await response.json();
        setProfile(data.profile);
        // bump avatarRefresh to bust browser cache when profile (and avatar) changes
        const ts = Date.now();
        setAvatarRefresh(ts);
        try { localStorage.setItem('avatarRefresh', String(ts)); } catch (e) { /* ignore */ }
        setPosts(data.posts || []);
        setFollowers(data.followers || []);
        setFollowing(data.following  || []);
        setPendingRequests(data.pendingRequests || []);
        setIsPrivate(data.profile?.isPrivate || false);
      }
    } catch (error) {
      console.error('Failed to fetch profile:', error);
    } finally {
      setLoading(false);
    }
  };
  const togglePrivacy = async () => {
    try {
      const response = await fetch('/api/toggle-privacy', {
        method: 'POST',
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        console.log('Privacy toggled:', data);
        setIsPrivate(data.isPrivate);
        // Refresh profile to get updated data
        fetchProfile();
      } else {
        const errorText = await response.text();
        console.error('Failed to toggle privacy:', response.status, errorText);
      }
    } catch (error) {
      console.error('Failed to toggle privacy:', error);
    }
  };

  useEffect(() => {
    if (user && user.userID) {
      fetchProfile();
    }
  }, [user?.userID]);

  // Subscribe to WebSocket events to refresh profile in real-time
  useEffect(() => {
    if (!user) return;
    const unsubscribe = subscribe((message) => {
      // Relevant message types that should refresh this profile view
      if (message.type === 'follow_request' || message.type === 'follow_request_response' || message.type === 'new_follower' || message.type === 'unfollow_update') {
        console.log('[MyProfile] WS event received, refreshing profile:', message.type);
        fetchProfile();
      }
    });

    return () => {
      unsubscribe();
    };
  }, [user, subscribe]);



  const handleEditProfile = () => {
    // Convert ISO date to yyyy-MM-dd format for date input
    let formattedDate = '';
    if (profile?.dateOfBirth) {
      try {
        const date = new Date(profile.dateOfBirth);
        formattedDate = date.toISOString().split('T')[0];
      } catch (e) {
        formattedDate = '';
      }
    }
    
    setEditForm({
      firstname: profile?.fname || '',
      lastname: profile?.lname || '',
      nickname: profile?.nickname || '',
      bio: profile?.bio || '',
      email: profile?.email || '',
      age: profile?.age || '',
      gender: profile?.gender || '',
      dateOfBirth: formattedDate
    });
    setEditMode(true);
  };

  const handleSaveProfile = async (e) => {
    e.preventDefault();
    try {
      let requestBody;
      let headers = {};
      
      // Check if there's a file to upload
      if (editForm.avatarFile && editForm.avatarFile.size > 0) {
        // Use FormData for file upload
        const formData = new FormData();
        Object.keys(editForm).forEach(key => {
          if (key !== 'avatarFile' && editForm[key] !== null && editForm[key] !== '') {
            formData.append(key, editForm[key]);
          }
        });
        formData.append('avatarInput', editForm.avatarFile);
        requestBody = formData;
        console.log('Sending FormData with avatar file');
      } else {
        // Use JSON for text-only updates (backend accepts this)
        const profileData = {};
        Object.keys(editForm).forEach(key => {
          if (key !== 'avatarFile' && editForm[key] !== null && editForm[key] !== '') {
            profileData[key] = editForm[key];
          }
        });
        requestBody = JSON.stringify(profileData);
        headers['Content-Type'] = 'application/json';
        console.log('Sending JSON (no file):', profileData);
      }
      
      const response = await fetch('/api/update-profile', {
        method: 'POST',
        headers: headers,
        body: requestBody,
        credentials: 'include'
      });

      if (response.ok) {
        setEditMode(false);
        fetchProfile(); // Refresh profile data
      } else {
        const errorData = await response.json().catch(() => ({}));
        console.error('Failed to update profile:', errorData);
      }
    } catch (error) {
      console.error('Error updating profile:', error);
    }
  };

  const handleCancelEdit = () => {
    setEditMode(false);
    setEditForm({
      firstname: '',
      lastname: '',
      nickname: '',
      bio: '',
      email: '',
      age: '',
      gender: '',
      dateOfBirth: '',
      avatarFile: null
    });
  };

  const handleFollowRequest = async (requesterId, action) => {
    try {
      const response = await fetch('/api/handle-follow-request', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          requester_id: requesterId,
          action: action
        })
      });

      if (response.ok) {
        // Refresh profile data to update pending requests and followers
        fetchProfile();
      }
    } catch (error) {
      console.error('Error handling follow request:', error);
    }
  };

  return (
    <Layout>
      <Head>
        <title>My Profile - Social Network</title>
      </Head>
      
      {loading ? (
        <div className="loading-container">
          <div className="loading-text">Loading profile...</div>
        </div>
      ) : (
        <div className="posts-layout">
          <div className="posts-main-content" style={{ maxWidth: '1000px', margin: '0 auto' }}>
            {/* Profile Header */}
            <div className="post-card" style={{ marginBottom: '1.5rem' }}>
              <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', flex: 1 }}>
                  <img
                    src={profile?.avatar || '/css/icon.png'}
                    alt="Profile Avatar"
                    className="post-avatar"
                    style={{ width: '96px', height: '96px', borderRadius: '50%' }}
                  />
                  <div style={{ flex: 1 }}>
                    <h1 style={{ fontSize: '1.875rem', fontWeight: 'bold', color: '#1f2937', marginBottom: '0.25rem' }}>
                      {profile?.fname} {profile?.lname}
                      {profile?.nickname && <span style={{ fontSize: '1.125rem', color: '#6b7280', fontWeight: 'normal', marginLeft: '0.5rem' }}>"{profile.nickname}"</span>}
                    </h1>
                    <p style={{ color: '#6b7280', marginBottom: '0.5rem' }}>@{profile?.username}</p>
                    
                    {/* Additional Profile Info */}
                    <div style={{ fontSize: '0.875rem', color: '#6b7280', marginTop: '0.5rem' }}>
                      {profile?.email && (
                        <div style={{ marginBottom: '0.25rem' }}>Email: {profile.email}</div>
                      )}
                      {profile?.age && profile?.gender && (
                        <div style={{ marginBottom: '0.25rem' }}>{profile.age} years old - {profile.gender}</div>
                      )}
                      {profile?.dateOfBirth && (
                        <div>Born {new Date(profile.dateOfBirth).toLocaleDateString()}</div>
                      )}
                    </div>
                    
                    <div style={{ display: 'flex', gap: '1.5rem', marginTop: '0.75rem' }}>
                      <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        <strong>{followers.length}</strong> Followers
                      </span>
                      <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        <strong>{following.length}</strong> Following
                      </span>
                      <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                        <strong>{posts.length}</strong> Posts
                      </span>
                    </div>
                  </div>
                </div>
                
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  <button
                    onClick={handleEditProfile}
                    className="icon-btn like-btn"
                    style={{ minWidth: '120px' }}
                  >
                    Edit Profile
                  </button>
                  <button
                    onClick={togglePrivacy}
                    className={isPrivate ? 'icon-btn dislike-btn' : 'icon-btn comment-btn'}
                    style={{ minWidth: '120px' }}
                  >
                    {isPrivate ? 'Private' : 'Public'}
                  </button>
                </div>
              </div>

              {profile?.bio && (
                <div style={{ borderTop: '1px solid #e5e7eb', paddingTop: '1rem' }}>
                  <h3 style={{ fontSize: '1.125rem', fontWeight: '600', marginBottom: '0.5rem', color: '#1f2937' }}>About</h3>
                  <p style={{ color: '#4b5563' }}>{profile.bio}</p>
                </div>
              )}
            </div>

            {/* Edit Profile Form */}
            {editMode && (
              <div className="post-card" style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.5rem', fontWeight: '600', marginBottom: '1rem', color: '#1f2937' }}>Edit Profile</h2>
                <form onSubmit={handleSaveProfile}>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        First Name *
                      </label>
                      <input
                        type="text"
                        value={editForm.firstname}
                        onChange={(e) => setEditForm({...editForm, firstname: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        required
                      />
                    </div>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Last Name *
                      </label>
                      <input
                        type="text"
                        value={editForm.lastname}
                        onChange={(e) => setEditForm({...editForm, lastname: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        required
                      />
                    </div>
                  </div>

                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Nickname (Optional)
                      </label>
                      <input
                        type="text"
                        value={editForm.nickname}
                        onChange={(e) => setEditForm({...editForm, nickname: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        placeholder="Your nickname"
                      />
                    </div>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Email Address *
                      </label>
                      <input
                        type="email"
                        value={editForm.email}
                        onChange={(e) => setEditForm({...editForm, email: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        required
                      />
                    </div>
                  </div>

                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Age *
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="120"
                        value={editForm.age}
                        onChange={(e) => setEditForm({...editForm, age: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        required
                      />
                    </div>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Gender *
                      </label>
                      <select
                        value={editForm.gender}
                        onChange={(e) => setEditForm({...editForm, gender: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                        required
                      >
                        <option value="">Select Gender</option>
                        <option value="male">Male</option>
                        <option value="female">Female</option>
                      </select>
                    </div>
                    <div>
                      <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                        Date of Birth
                      </label>
                      <input
                        type="date"
                        value={editForm.dateOfBirth}
                        onChange={(e) => setEditForm({...editForm, dateOfBirth: e.target.value})}
                        style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem' }}
                      />
                    </div>
                  </div>
                  
                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '500', color: '#374151', marginBottom: '0.25rem' }}>
                      Bio
                    </label>
                    <textarea
                      value={editForm.bio}
                      onChange={(e) => setEditForm({...editForm, bio: e.target.value})}
                      style={{ width: '100%', padding: '0.5rem 0.75rem', border: '1px solid #d1d5db', borderRadius: '0.5rem', resize: 'vertical' }}
                      rows={3}
                      placeholder="Tell us about yourself..."
                    />
                  </div>

                  <div style={{ marginBottom: '1rem' }}>
                    <label style={{ display: 'block', fontSize: '0.875rem', fontWeight: '600', color: '#374151', marginBottom: '0.75rem' }}>
                      Profile Picture (Optional)
                    </label>
                    
                    {!editForm.avatarFile ? (
                      <label 
                        htmlFor="avatar-upload"
                        style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          gap: '1rem', 
                          padding: '1rem',
                          border: '2px dashed #d1d5db',
                          borderRadius: '1rem',
                          background: 'linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%)',
                          cursor: 'pointer',
                          transition: 'all 0.3s ease'
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.borderColor = '#8b5cf6';
                          e.currentTarget.style.background = 'linear-gradient(135deg, #faf5ff 0%, #f3e8ff 100%)';
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.borderColor = '#d1d5db';
                          e.currentTarget.style.background = 'linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%)';
                        }}
                      >
                        <div style={{ flexShrink: 0 }}>
                          <img
                            src={profile?.avatar || '/css/icon.png'}
                            alt="Current avatar"
                            style={{ width: '64px', height: '64px', borderRadius: '50%', border: '3px solid white', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }}
                          />
                        </div>
                        <div style={{ flex: 1 }}>
                          <div style={{ fontWeight: '600', color: '#374151', marginBottom: '0.25rem', fontSize: '0.9rem' }}>
                            Click to change profile picture
                          </div>
                          <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
                            JPEG, PNG, GIF • Max 10MB
                          </div>
                        </div>
                        <div style={{ flexShrink: 0 }}>
                          <svg style={{ width: '24px', height: '24px', color: '#8b5cf6' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
                          </svg>
                        </div>
                        <input
                          id="avatar-upload"
                          type="file"
                          accept="image/jpeg,image/png,image/gif"
                          onChange={(e) => setEditForm({...editForm, avatarFile: e.target.files[0]})}
                          style={{ display: 'none' }}
                        />
                      </label>
                    ) : (
                      <div 
                        style={{ 
                          padding: '1rem', 
                          borderRadius: '1rem', 
                          background: 'linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%)',
                          border: '2px solid #10b981'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div style={{ flexShrink: 0 }}>
                            <div style={{ width: '64px', height: '64px', borderRadius: '50%', overflow: 'hidden', border: '3px solid white', boxShadow: '0 4px 6px rgba(0,0,0,0.1)' }}>
                              <img
                                src={URL.createObjectURL(editForm.avatarFile)}
                                alt="New avatar preview"
                                style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                              />
                            </div>
                          </div>
                          <div style={{ flex: 1 }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
                              <svg style={{ width: '20px', height: '20px', color: '#10b981' }} fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                              </svg>
                              <span style={{ fontWeight: '600', color: '#047857', fontSize: '0.9rem' }}>New picture selected!</span>
                            </div>
                            <div style={{ fontSize: '0.75rem', color: '#059669' }}>
                              {editForm.avatarFile.name}
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() => setEditForm({...editForm, avatarFile: null})}
                            style={{ 
                              flexShrink: 0,
                              padding: '0.5rem 1rem',
                              background: 'white',
                              border: '2px solid #ef4444',
                              borderRadius: '0.5rem',
                              color: '#ef4444',
                              fontWeight: '600',
                              fontSize: '0.875rem',
                              cursor: 'pointer',
                              transition: 'all 0.2s'
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.style.background = '#fef2f2';
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.style.background = 'white';
                            }}
                          >
                            Remove
                          </button>
                        </div>
                      </div>
                    )}
                  </div>

                  <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <button
                      type="submit"
                      className="icon-btn like-btn"
                      style={{ minWidth: '120px' }}
                    >
                      Save Changes
                    </button>
                    <button
                      type="button"
                      onClick={handleCancelEdit}
                      className="icon-btn"
                      style={{ minWidth: '120px', backgroundColor: '#e5e7eb', color: '#374151' }}
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              </div>
            )}

            {/* Pending Follow Requests */}
            {pendingRequests.length > 0 && (
              <div className="post-card" style={{ marginBottom: '1.5rem' }}>
                <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#1f2937', marginBottom: '1rem' }}>
                  Follow Requests ({pendingRequests.length})
                </h2>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  {pendingRequests.map((request) => (
                    <div key={request.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.75rem', backgroundColor: '#f9fafb', borderRadius: '0.5rem' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                        <img
                          src={request.avatar || '/css/icon.png'}
                          alt={request.username}
                          style={{ width: '40px', height: '40px', borderRadius: '50%' }}
                        />
                        <div>
                          <div style={{ fontWeight: '500', color: '#1f2937' }}>
                            {request.firstname} {request.lastname}
                          </div>
                          <div style={{ fontSize: '0.875rem', color: '#6b7280' }}>@{request.username}</div>
                          <div style={{ fontSize: '0.75rem', color: '#9ca3af' }}>
                            {new Date(request.created_at).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button
                          onClick={() => handleFollowRequest(request.id, 'accept')}
                          className="icon-btn like-btn"
                          style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}
                        >
                          Accept
                        </button>
                        <button
                          onClick={() => handleFollowRequest(request.id, 'decline')}
                          className="icon-btn dislike-btn"
                          style={{ fontSize: '0.875rem', padding: '0.5rem 1rem' }}
                        >
                          Decline
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* My Posts */}
            <div className="post-card" style={{ marginBottom: '1.5rem' }}>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#1f2937', marginBottom: '1rem' }}>My Posts ({posts.length})</h2>
              
              {posts.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '2rem' }}>
                  <p style={{ color: '#6b7280', marginBottom: '1rem' }}>You haven't created any posts yet</p>
                  <button
                    onClick={() => router.push('/createPost')}
                    className="icon-btn like-btn"
                  >
                    Create Your First Post
                  </button>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {posts.map((post) => (
                    <div key={post.id} className="post-card" style={{ padding: '1rem' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '0.5rem' }}>
                        <h3 style={{ fontSize: '1.125rem', fontWeight: '600', color: '#1f2937' }}>{post.title}</h3>
                        <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                          {(() => {
                            let dateVal = post.createdAt || post.created_at;
                            if (dateVal) {
                              // If it's a number (timestamp), parse as int
                              if (typeof dateVal === 'number' || /^\d+$/.test(dateVal)) {
                                const date = new Date(Number(dateVal));
                                return isNaN(date) ? 'Unknown Date' : date.toLocaleDateString();
                              }
                              // If it's a string, try parsing as ISO/date string
                              const date = new Date(dateVal);
                              return isNaN(date) ? 'Unknown Date' : date.toLocaleDateString();
                            }
                            return 'Unknown Date';
                          })()}
                        </span>
                      </div>
                      
                      <p style={{ color: '#4b5563', marginBottom: '0.75rem' }}>{post.content}</p>
                      
                      {post.category && (
                        <span style={{ display: 'inline-block', backgroundColor: '#f3f4f6', color: '#374151', padding: '0.25rem 0.75rem', borderRadius: '9999px', fontSize: '0.875rem', marginBottom: '0.75rem' }}>
                          {post.category}
                        </span>
                      )}
                      
                      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', paddingTop: '0.75rem', borderTop: '1px solid #e5e7eb' }}>
                        {/* <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                          Likes: {post.likes || 0}
                        </span> */}
                        {/* <span style={{ fontSize: '0.875rem', color: '#6b7280' }}>
                          Comments: {post.comments || 0}
                        </span> */}
                        <button
                          onClick={() => router.push(`/comments?postId=${post.id}`)}
                          className="icon-btn comment-btn"
                          style={{ fontSize: '0.875rem', padding: '0.25rem 0.75rem', marginLeft: 'auto' }}
                        >
                          View Post
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Followers and Following */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
              {/* Followers */}
              <div className="post-card">
                <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#1f2937', marginBottom: '1rem' }}>
                  Followers ({followers.length})
                </h2>
                {followers.length === 0 ? (
                  <p style={{ color: '#6b7280', textAlign: 'center', padding: '1rem' }}>No followers yet</p>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', maxHeight: '400px', overflowY: 'auto' }}>
                    {followers.map((follower) => (
                      <div key={follower.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', borderRadius: '0.375rem', transition: 'background-color 0.2s' }} onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'} onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                          <img
                            src={follower.avatar || '/css/icon.png'}
                            alt={follower.username}
                            style={{ width: '32px', height: '32px', borderRadius: '50%' }}
                          />
                          <div>
                            <div style={{ fontWeight: '500', color: '#1f2937', fontSize: '0.875rem' }}>
                              {follower.firstname} {follower.lastname}
                            </div>
                            <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>@{follower.username}</div>
                          </div>
                        </div>
                        <button
                          onClick={() => startChat(follower)}
                          className="icon-btn comment-btn"
                          style={{ fontSize: '0.875rem', padding: '0.25rem 0.5rem' }}
                        >
                          Chat
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Following */}
              <div className="post-card">
                <h2 style={{ fontSize: '1.25rem', fontWeight: 'bold', color: '#1f2937', marginBottom: '1rem' }}>
                  Following ({following.length})
                </h2>
                {following.length === 0 ? (
                  <p style={{ color: '#6b7280', textAlign: 'center', padding: '1rem' }}>Not following anyone yet</p>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', maxHeight: '400px', overflowY: 'auto' }}>
                    {following.map((user) => (
                      <div key={user.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', borderRadius: '0.375rem', transition: 'background-color 0.2s' }} onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#f9fafb'} onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                          <img
                            src={user.avatar || '/css/icon.png'}
                            alt={user.username}
                            style={{ width: '32px', height: '32px', borderRadius: '50%' }}
                          />
                          <div>
                            <div style={{ fontWeight: '500', color: '#1f2937', fontSize: '0.875rem' }}>
                              {user.firstname} {user.lastname}
                            </div>
                            <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>@{user.username}</div>
                          </div>
                        </div>
                        <button
                          onClick={() => startChat(user)}
                          className="icon-btn comment-btn"
                          style={{ fontSize: '0.875rem', padding: '0.25rem 0.5rem' }}
                        >
                          Chat
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

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
