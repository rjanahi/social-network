import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Layout from '../components/Layout';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080';

export default function Comments() {
  const router = useRouter();
  const { postId, type, groupId } = router.query;
  const { user, loading: sessionLoading } = useSession();
  const { subscribe } = useWebSocketContext();
  const isGroupPost = type === 'group';
  const [post, setPost] = useState(null);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [commentImage, setCommentImage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (postId) {
      fetchCommentsAndPost();
    }
  }, [postId]);

  // Listen for real-time comment updates
  useEffect(() => {
    if (!user || !postId) return;

    const unsubscribe = subscribe((message) => {
      // Listen for both regular and group post comments
      if (message.type === 'new_comment' && message.post_id == postId) {
        fetchCommentsAndPost();
      } else if (message.type === 'group_post_comment' && message.post_id == postId) {
        try {
          const commentData = JSON.parse(message.content);
          setComments(prev => {
            // Avoid duplicates
            if (prev.some(c => c.id === commentData.id)) return prev;
            return [...prev, commentData];
          });
        } catch (e) {
          console.error('[Comments] Failed to parse new group comment data:', e);
        }
      } else if (message.type === 'new_commentLike') {
        try {
          const likeData = JSON.parse(message.content);
          setComments(prev => prev.map(comment => 
            comment.id === likeData.comment_id 
              ? { 
                  ...comment, 
                  likes_count: likeData.likes_count,
                  dislikes_count: likeData.dislikes_count 
                }
              : comment
          ));
        } catch (e) {
          console.error('[Comments] Failed to parse comment like data:', e);
        }
      }
    });

    return () => unsubscribe();
  }, [user, postId, subscribe]);

  const fetchCommentsAndPost = async () => {
    try {
      setLoading(true);
      setError(null);
      
      if (isGroupPost && groupId) {
        // For group posts, fetch directly from backend
        const response = await fetch(`${BACKEND_URL}/groups/${groupId}/posts/${postId}/comments`, {
          credentials: 'include'
        });
        
        if (response.ok) {
          const data = await response.json();
          // Backend now returns { post: {...}, comments: [...] }
          setPost(data.post);
          setComments(data.comments || []);
        } else {
          setError('Failed to fetch group comments');
        }
      } else {
        // Regular posts
        const response = await fetch(`/api/comments/${postId}`, {
          credentials: 'include'
        });
        
        if (response.ok) {
          const data = await response.json();
          setPost(data.post);
          setComments(data.comments || []);
        } else {
          setError('Failed to fetch comments');
        }
      }
    } catch (error) {
      console.error('Error fetching comments:', error);
      setError('Error loading comments');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitComment = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    try {
      let response;
      
      if (isGroupPost && groupId) {
        // For group posts, use backend API with FormData
        const formData = new FormData();
        formData.append('group_id', groupId);
        formData.append('group_post_id', postId);
        formData.append('content', newComment.trim());
        if (commentImage) {
          formData.append('imgOrgif', commentImage);
        }
        
        response = await fetch(`${BACKEND_URL}/groups/comments`, {
          method: 'POST',
          body: formData,
          credentials: 'include'
        });
      } else {
        // Regular posts - use FormData for image support
        const formData = new FormData();
        formData.append('post_id', postId);
        formData.append('comment', newComment.trim());
        if (commentImage) {
          formData.append('imgOrgif', commentImage);
        }

        response = await fetch('/api/comment', {
          method: 'POST',
          body: formData,
          credentials: 'include'
        });
      }

      if (response.ok) {
        setNewComment('');
        setCommentImage(null);
        fetchCommentsAndPost(); // Refresh comments
      } else {
        console.error('Failed to post comment');
      }
    } catch (error) {
      console.error('Error posting comment:', error);
    }
  };

  // Listen for ReplaceButton file selection (custom event) for comment image
  useEffect(() => {
    const onReplace = (ev) => {
      const file = ev.detail;
      if (file) setCommentImage(file);
    };
    window.addEventListener('replace-image-selected', onReplace);
    return () => window.removeEventListener('replace-image-selected', onReplace);
  }, []);

  if (loading || sessionLoading) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Loading comments...</div>
        </div>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="error-text">{error}</div>
        </div>
      </Layout>
    );
  }

  if (!post) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Post not found</div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <Head>
        <title>Comments - {post.title}</title>
      </Head>
      
      <div className="comments-container">
        <button 
          onClick={() => router.back()}
          className="back-button"
        >
          ← Back to Posts
        </button>

        {/* Original Post */}
        <div className="post-detail">
          <div className="post-header">
            {post.avatar_url && post.avatar_url !== '' ? (
              <img
                src={post.avatar_url}
                alt={`${post.username || 'User'}'s avatar`}
                className="post-avatar"
                onError={e => { e.target.src = '/img/avatars/images.png'; }}
              />
            ) : (
              <div
                className="post-avatar-fallback"
                style={{
                  width: '48px',
                  height: '48px',
                  borderRadius: '50%',
                  background: '#e5e7eb',
                  color: '#374151',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontWeight: 700,
                  fontSize: '1.5rem',
                  border: '2px solid #e5e7eb',
                  textTransform: 'uppercase',
                }}
              >
                {post.firstname && post.firstname.length > 0
                  ? post.firstname[0]
                  : (post.username && post.username.length > 0
                      ? post.username[0]
                      : '?')}
              </div>
            )}
            <div className="post-user-info">
              <h3>{post.username || 'Anonymous'}</h3>
              <p className="post-date">
                {(() => {
                  const d = new Date(post.created_at || post.createdAt);
                  return isNaN(d.getTime()) ? (post.created_at || post.createdAt || 'Unknown Date') : d.toLocaleString();
                })()}
              </p>
            </div>
          </div>
          
          <h1 className="post-title">{post.title || 'Untitled Post'}</h1>
          <p className="post-content">{post.content || 'No content available'}</p>
          
          {post.image && (
            <img
              src={post.image}
              alt="Post image"
              className="post-image"
            />
          )}
        </div>

        {/* Add Comment Form */}
        <div className="comment-form">
          <h2>Add a Comment</h2>
          <form onSubmit={handleSubmitComment}>
            <textarea
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              placeholder="Write your comment here..."
              className="comment-textarea"
              rows={3}
              required
            />
            
            {/* Image Upload */}
            <div className="mt-3">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Add Image or GIF (Optional)
              </label>

              {/* Use the same modern compact upload box as Create Post */}
              {!commentImage ? (
                <div className="relative">
                  <label
                    htmlFor="comment-image"
                    className="group flex items-center gap-4 p-4 rounded-2xl border-2 border-dashed cursor-pointer transition-all duration-300 hover:shadow-md"
                    style={{ background: 'var(--primary-50)', borderColor: 'var(--primary-200)' }}
                  >
                    <div className="flex-shrink-0">
                      <div className="w-12 h-12 rounded-xl flex items-center justify-center shadow-lg group-hover:scale-105 transition-transform duration-200" style={{ background: 'linear-gradient(135deg, var(--primary-500), var(--primary-600))' }}>
                        <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 4v16m8-8H4" />
                        </svg>
                      </div>
                    </div>

                    <div className="flex-1">
                      <h4 className="text-md font-bold mb-1 transition-colors" style={{ color: '#374151' }}>
                        Click to upload an image
                      </h4>
                      <p className="text-sm font-medium" style={{ color: '#6b7280' }}>
                        Supports: JPEG, PNG, GIF (Max 10MB)
                      </p>
                    </div>

                    <div className="flex-shrink-0 flex items-center gap-3">
                      <svg className="w-5 h-5" style={{ color: '#667eea' }} viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </label>

                  <input
                    type="file"
                    id="comment-image"
                    accept="image/jpeg,image/jpg,image/png,image/gif"
                    onChange={(e) => { if (e.target.files && e.target.files[0]) setCommentImage(e.target.files[0]); }}
                    className="sr-only"
                  />
                </div>
              ) : (
                <div className="bg-gradient-to-br from-green-50 to-emerald-50 rounded-2xl p-4 border-2 border-green-300 shadow relative overflow-hidden">
                  <div className="flex items-start gap-3">
                    <div className="flex-shrink-0 relative group">
                      <div className="w-20 h-20 rounded-xl overflow-hidden shadow-md border-2 border-white">
                        <img src={URL.createObjectURL(commentImage)} alt="Preview" className="w-full h-full object-cover" />
                      </div>
                      <div className="absolute inset-0 bg-black/0 group-hover:bg-black/8 rounded-xl transition-colors duration-200"></div>
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div className="flex items-center gap-2">
                          <svg className="w-5 h-5 text-green-600 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                          </svg>
                        </div>
                      </div>

                      <p className="text-sm text-green-700 font-medium mb-3 truncate">{commentImage.name}</p>

                      <div className="flex gap-2 image-action-group">
                        <button
                          type="button"
                          onClick={() => setCommentImage(null)}
                          className="inline-flex items-center gap-2 px-4 py-2 bg-white border-2 border-red-300 text-red-700 rounded-xl font-semibold text-sm hover:bg-red-50 hover:border-red-400 transition-all duration-150 shadow-sm image-action-btn image-action-remove"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                          Remove
                        </button>

                        <ReplaceButton />
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
            
            <div className="comment-form-actions">
              <button
                type="submit"
                className="btn btn-primary"
              >
                Post Comment
              </button>
            </div>
          </form>
        </div>

        {/* Comments List */}
        <div className="comments-list">
          <h2 className="comments-section-title">
            Comments ({comments.length})
          </h2>
          
          {comments.length === 0 ? (
            <div className="no-comments">
              <p>No comments yet. Be the first to comment!</p>
            </div>
          ) : (
            comments.map((comment) => (
              <div key={comment.id} className="comment-card">
                <div className="comment-header">
                  {comment.avatar_url && comment.avatar_url !== '' ? (
                    <img
                      src={comment.avatar_url}
                      alt={`${comment.username}'s avatar`}
                      className="comment-avatar"
                      onError={e => { e.target.src = '/img/avatars/images.png'; }}
                    />
                  ) : (
                    <div
                      className="comment-avatar-fallback"
                      style={{
                        width: '40px',
                        height: '40px',
                        borderRadius: '50%',
                        background: '#e5e7eb',
                        color: '#374151',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontWeight: 700,
                        fontSize: '1.2rem',
                        border: '2px solid #e5e7eb',
                        textTransform: 'uppercase',
                      }}
                    >
                      {comment.firstname && comment.firstname.length > 0
                        ? comment.firstname[0]
                        : (comment.username && comment.username.length > 0
                            ? comment.username[0]
                            : '?')}
                    </div>
                  )}
                  <div className="comment-user-info">
                    <h4>{comment.username || 'Anonymous'}</h4>
                    <p className="post-date">
                      {(() => {
                        const d = new Date(comment.created_at || comment.createdAt);
                        return isNaN(d.getTime()) ? (comment.created_at || comment.createdAt || 'Unknown Date') : d.toLocaleString();
                      })()}
                    </p>
                  </div>
                </div>
                <p className="comment-content">{comment.content}</p>
                
                {/* Display comment image if exists */}
                {comment.image && (
                  <img
                    src={comment.image}
                    alt="Comment image"
                    className="comment-image mt-2 max-w-md rounded-lg border border-gray-200"
                  />
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </Layout>
  );
}

function ReplaceButton() {
  const replaceRef = useRef(null);

  return (
    <>
      <button
        type="button"
        onClick={() => replaceRef.current && replaceRef.current.click()}
        className="inline-flex items-center gap-2 px-4 py-2 bg-white border-2 border-purple-300 text-purple-700 rounded-xl font-semibold text-sm hover:bg-purple-50 hover:border-purple-400 transition-all duration-200 shadow-sm hover:shadow cursor-pointer image-action-btn image-action-replace"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Replace
      </button>
      <input
        type="file"
        id="comment-image-replace"
        name="comment-image-replace"
        accept="image/jpeg,image/jpg,image/png,image/gif"
        onChange={(e) => {
          if (e.target.files && e.target.files[0]) {
            const file = e.target.files[0];
            const custom = new CustomEvent('replace-image-selected', { detail: file });
            window.dispatchEvent(custom);
          }
        }}
        ref={replaceRef}
        className="sr-only"
      />
    </>
  );
}
