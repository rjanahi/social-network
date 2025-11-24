import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/router';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import Toast from './Toast';

export default function GroupPostsSection({ groupId, user }) {
  const [posts, setPosts] = useState([]);
  const [loadingPosts, setLoadingPosts] = useState(true);
  const [showCreatePost, setShowCreatePost] = useState(false);
  const [newPost, setNewPost] = useState({ title: '', content: '' });
  const [postImage, setPostImage] = useState(null);
  const router = useRouter();
  const { subscribe } = useWebSocketContext();
  const [toast, setToast] = useState({ visible: false, message: '', type: 'info' });

  const showToast = (message, type = 'info', duration = 3500) => {
    setToast({ visible: true, message, type, duration });
  };

  const hideToast = () => setToast(prev => ({ ...prev, visible: false }));

  // Debug log to check groupId
  console.log('GroupPostsSection received groupId:', groupId, 'type:', typeof groupId);

  useEffect(() => {
    if (groupId) {
      fetchGroupPosts();
    }
  }, [groupId]);

  // Listen for real-time post updates
  useEffect(() => {
    if (!user || !groupId) return;

    const unsubscribe = subscribe((message) => {
      console.log('[GroupPostsSection] Received WebSocket message:', message);
      if (message.group_id == groupId) {
        if (message.type === 'new_groupPost') {
          try {
            const postData = JSON.parse(message.content);
            setPosts(prev => {
              // Avoid duplicates
              if (prev.some(p => p.id === postData.id)) return prev;
              return [postData, ...prev];
            });
          } catch (e) {
            console.error('[GroupPostsSection] Failed to parse new group post data:', e);
          }
        } else if (message.type === 'group_post_like_update') {
          try {
            const likeData = JSON.parse(message.content);
            setPosts(prev => prev.map(post => 
              post.id === likeData.post_id 
                ? { 
                    ...post, 
                    like_count: likeData.likes_count,
                    dislike_count: likeData.dislikes_count 
                  }
                : post
            ));
          } catch (e) {
            console.error('[GroupPostsSection] Failed to parse like data:', e);
          }
        } else if (message.type === 'group_post_comment') {
          fetchGroupPosts();
        }
      }
    });

    return () => unsubscribe();
  }, [user, groupId, subscribe]);

  const fetchGroupPosts = async () => {
    if (!groupId) return;
    
    try {
      setLoadingPosts(true);
      console.log('Fetching group posts for groupId:', groupId);
      
      const response = await fetch(`http://localhost:8080/groups/${groupId}/posts`, {
        method: 'GET',
        credentials: 'include'
      });
      
      console.log('Response status:', response.status);
      
      if (response.ok) {
        const postsData = await response.json();
        console.log('Posts data received:', postsData);
        setPosts(Array.isArray(postsData) ? postsData : []);
      } else {
        const errorText = await response.text();
        console.error('Failed to fetch group posts:', response.status, errorText);
        setPosts([]);
      }
    } catch (error) {
      console.error('Error fetching group posts:', error);
      setPosts([]);
    } finally {
      setLoadingPosts(false);
    }
  };

  const handleCreatePost = async (e) => {
    e.preventDefault();
    if (!newPost.title.trim() || !newPost.content.trim()) {
      showToast('Please fill in both title and content', 'error');
      return;
    }

    if (!groupId) {
      showToast(`Group ID is missing. Received: ${groupId} (type: ${typeof groupId})`, 'error');
      return;
    }
    
    try {
      const formData = new FormData();
      formData.append('group_id', groupId.toString());
      formData.append('title', newPost.title.trim());
      formData.append('content', newPost.content.trim());
      
      // Add image if selected
      if (postImage) {
        formData.append('imgOrgif', postImage);
      }
      
      console.log('Creating post with groupId:', groupId, 'FormData:', {
        group_id: groupId.toString(),
        title: newPost.title.trim(),
        content: newPost.content.trim(),
        hasImage: !!postImage
      });

      // Debug: Log FormData entries
      for (let pair of formData.entries()) {
        console.log('FormData entry:', pair[0], '=', pair[1]);
      }
      
      const response = await fetch('http://localhost:8080/groups/create-post', {
        method: 'POST',
        body: formData,
        credentials: 'include'
      });
      
      if (response.ok) {
        setShowCreatePost(false);
        setNewPost({ title: '', content: '' });
        setPostImage(null);
  fetchGroupPosts(); // Refresh posts
      } else {
        const errorText = await response.text();
        showToast(`Failed to create post: ${errorText}`, 'error');
      }
    } catch (error) {
      console.error('Error creating post:', error);
      showToast('Error creating post', 'error');
    }
  };

  const handleLike = async (postId) => {
    try {
      const response = await fetch(`http://localhost:8080/groups/likes?groupId=${groupId}&postId=${postId}`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        // Update local state immediately with the response data
        setPosts(prev => prev.map(post => 
          post.id === postId 
            ? { 
                ...post, 
                like_count: data.likes,
                dislike_count: data.dislikes,
                is_liked: !post.is_liked  // Toggle the liked state
              }
            : post
        ));
      } else {
        console.error('Failed to toggle like');
      }
    } catch (error) {
      console.error('Error toggling like:', error);
    }
  };

  const handleDislike = async (postId) => {
    try {
      const response = await fetch(`http://localhost:8080/groups/dislikes?groupId=${groupId}&postId=${postId}`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        // Update local state immediately with the response data
        setPosts(prev => prev.map(post => 
          post.id === postId 
            ? { 
                ...post, 
                like_count: data.likes,
                dislike_count: data.dislikes,
                is_disliked: !post.is_disliked  // Toggle the disliked state
              }
            : post
        ));
      } else {
        console.error('Failed to toggle dislike');
      }
    } catch (error) {
      console.error('Error toggling dislike:', error);
    }
  };

  return (
    <>
      <div className="group-posts-section-modern">
      {/* Modern Create Post Prompt */}
      <div className="create-post-prompt-modern">
        <button
          onClick={() => setShowCreatePost(true)}
          className="create-post-button-modern"
        >
          <div className="create-post-avatar">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <circle cx="12" cy="7" r="4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
          <span className="create-post-placeholder">What's on your mind? Share with your group...</span>
          <div className="create-post-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
              <path d="M8.5 10a1.5 1.5 0 100-3 1.5 1.5 0 000 3zM21 15l-5-5L5 21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
        </button>
      </div>

      {/* Create Post Modal */}
      {showCreatePost && (
        <div className="modal">
          <div className="modal-content">
            <span className="close" onClick={() => setShowCreatePost(false)}>&times;</span>
            <h2>Create New Group Post</h2>
            
            <form onSubmit={handleCreatePost}>
              <div className="form-group">
                <label>Post Title</label>
                <input
                  type="text"
                  value={newPost.title}
                  onChange={(e) => setNewPost({...newPost, title: e.target.value})}
                  placeholder="Enter a catchy title..."
                  required
                />
              </div>
              
              <div className="form-group">
                <label>Content</label>
                <textarea
                  value={newPost.content}
                  onChange={(e) => setNewPost({...newPost, content: e.target.value})}
                  rows="4"
                  placeholder="Share your thoughts with the group..."
                  required
                />
              </div>
              
              {/* Image Upload - modern card matching Create Post */}
              <div className="form-group">
                <label>Add Image or GIF (Optional)</label>

                {!postImage ? (
                  <div className="relative">
                    <label
                      htmlFor="group-post-image"
                      className="group flex flex-col items-center gap-3 p-4 rounded-2xl border-2 border-dashed cursor-pointer transition-all duration-150"
                      style={{ background: 'var(--primary-50)', borderColor: 'var(--primary-200)', maxWidth: '460px', minHeight: '110px' }}
                    >

                      <div style={{ textAlign: 'center' }}>
                        <div className="text-sm font-semibold" style={{ color: '#374151' }}>Upload image</div>
                        <div className="text-xs" style={{ color: '#6b7280', marginTop: '4px' }}>JPEG, PNG, GIF — max 10MB</div>
                      </div>
                    </label>

                    <input
                      type="file"
                      id="group-post-image"
                      accept="image/jpeg,image/jpg,image/png,image/gif"
                      onChange={(e) => { if (e.target.files && e.target.files[0]) setPostImage(e.target.files[0]); }}
                      className="sr-only"
                    />
                  </div>
                ) : (
                  <div className="rounded-2xl p-3 border shadow-sm relative overflow-hidden" style={{ background: 'var(--primary-50)', borderColor: 'var(--primary-200)' }}>
                    <div className="flex items-center gap-3">
                      <div className="flex-shrink-0 relative">
                        <div className="w-12 h-12 rounded-md overflow-hidden shadow-sm border-2 border-white">
                          <img src={URL.createObjectURL(postImage)} alt="Preview" className="w-full h-full object-cover" />
                        </div>
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-2 mb-0">
                          <p className="text-sm text-gray-700 font-medium truncate" style={{ margin: 0, color: 'var(--primary-700)' }}>{postImage.name}</p>
                        </div>

                        <div className="flex gap-2 image-action-group" style={{ marginTop: '6px' }}>
                          <button
                            type="button"
                            onClick={() => setPostImage(null)}
                            className="inline-flex items-center gap-2 px-4 py-2 bg-white border-2 border-purple-200 text-purple-700 rounded-xl font-semibold text-sm hover:bg-purple-50 hover:border-purple-300 transition-all duration-150 shadow-sm image-action-btn image-action-remove"
                          >
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                            Remove
                          </button>

                          <GroupReplaceButton setPostImage={setPostImage} />
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
              
              <div className="form-group">
                <button type="button" onClick={() => setShowCreatePost(false)} className="btn btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Share Post
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Posts List */}
      {loadingPosts ? (
        <div className="loading-container">
          <div className="loading-text">Loading posts...</div>
        </div>
      ) : posts.length === 0 ? (
        <div className="no-posts-message">
          <div style={{fontSize: '3rem', marginBottom: '1rem'}}>📝</div>
          <h3>No posts yet</h3>
          <p>Be the first to share something with your group!</p>
          <button
            onClick={() => setShowCreatePost(true)}
            className="btn btn-primary"
          >
            Create First Post
          </button>
        </div>
      ) : (
        <div className="posts-grid-modern">
          {posts.map(post => (
            <div key={post.id} className="post-card-modern">
              {/* Post Header */}
              <div className="post-header-modern">
                <div className="post-author-modern">
                  {post.avatar_url && post.avatar_url !== '' ? (
                    <img
                      src={post.avatar_url}
                      alt={`${post.username || 'User'}'s avatar`}
                      className="post-avatar-modern"
                      onError={e => { e.target.src = '/img/avatars/images.png'; }}
                    />
                  ) : (
                    <div className="post-avatar-fallback-modern">
                      {post.firstname && post.firstname.length > 0
                        ? post.firstname[0]
                        : (post.username && post.username.length > 0
                            ? post.username[0]
                            : '?')}
                    </div>
                  )}
                  <div className="post-author-info-modern">
                    <h3 className="post-author-name">{post.username || 'Unknown User'}</h3>
                    <div className="post-meta-modern">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                        <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                      </svg>
                      {new Date(post.created_at).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit'
                      })}
                    </div>
                  </div>
                </div>
              </div>

              {/* Post Content */}
              <div className="post-body-modern">
                <h2 className="post-title-modern">{post.title}</h2>
                <p className="post-content-modern">{post.content}</p>

                {/* Post Image */}
                {post.image && (
                  <div className="post-image-modern">
                    <img src={post.image} alt="Post image" />
                  </div>
                )}
              </div>

              {/* Post Actions */}
              <div className="post-actions-modern">
                <button
                  onClick={() => handleLike(post.id)}
                  className={`post-action-btn ${post.is_liked ? 'active liked' : ''}`}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M7 22V11M2 13v6a2 2 0 002 2h2.4M2 13l6.5-6.5M2 13h6M22 13l-4-9H7l4 9h11zM18 13l-2.5 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span>{post.like_count || 0}</span>
                </button>
                
                <button
                  onClick={() => handleDislike(post.id)}
                  className={`post-action-btn ${post.is_disliked ? 'active disliked' : ''}`}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M17 2v11M22 11V5a2 2 0 00-2-2h-2.4M22 11l-6.5 6.5M22 11h-6M2 11l4 9h11l-4-9H2zM6 11l2.5-9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span>{post.dislike_count || 0}</span>
                </button>
                
                <button
                  onClick={() => router.push(`/comments?postId=${post.id}&type=group&groupId=${groupId}`)}
                  className="post-action-btn"
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2v10z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span>{post.comment_count || 0}</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
      </div>
      <Toast visible={toast.visible} message={toast.message} type={toast.type} onClose={hideToast} duration={toast.duration} />
    </>
  );
}

function GroupReplaceButton({ setPostImage }) {
  const fileRef = useRef(null);

  const onReplaceClick = () => { if (fileRef.current) fileRef.current.click(); };

  const onFileChange = (e) => { if (e.target.files && e.target.files[0]) setPostImage(e.target.files[0]); };

  return (
    <>
      <button
        type="button"
        onClick={onReplaceClick}
        className="inline-flex items-center gap-2 px-4 py-2 bg-white border-2 border-purple-300 text-purple-700 rounded-xl font-semibold text-sm hover:bg-purple-50 hover:border-purple-400 transition-all duration-200 shadow-sm hover:shadow cursor-pointer image-action-btn image-action-replace"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Replace
      </button>
      <input
        ref={fileRef}
        type="file"
        accept="image/jpeg,image/jpg,image/png,image/gif"
        onChange={onFileChange}
        className="sr-only"
      />
    </>
  );
}