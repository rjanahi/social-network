import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Layout from '../components/Layout';
import { useSession } from '../hooks/useSession';

export default function CreatePost() {
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    category: '',
    privacy: 'public',
    image: null,
    selectedFollowers: []
  });
  const [followers, setFollowers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { user, loading: sessionLoading } = useSession();
  const router = useRouter();

  const categories = ['Travel', 'Sport', 'Food', 'Nature'];
  const privacyOptions = [
    { value: 'public', label: 'Public', description: 'Everyone can see' },
    { value: 'almost-private', label: 'Almost Private', description: 'Only followers' },
    { value: 'private', label: 'Private', description: 'Selected followers only' }
  ];

  // Fetch followers when privacy is set to private
  useEffect(() => {
    if (formData.privacy === 'private') {
      fetchFollowers();
    }
  }, [formData.privacy]);

  const fetchFollowers = async () => {
    try {
      const response = await fetch('/api/user-followers', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        setFollowers(data.followers || []);
      }
    } catch (error) {
      console.error('Error fetching followers:', error);
    }
  };

  const handleChange = (e) => {
    if (e.target.type === 'file') {
      setFormData({
        ...formData,
        [e.target.name]: e.target.files[0]
      });
    } else {
      setFormData({
        ...formData,
        [e.target.name]: e.target.value
      });
    }
  };

  const handleFollowerToggle = (followerId) => {
    setFormData(prev => {
      const selected = prev.selectedFollowers.includes(followerId)
        ? prev.selectedFollowers.filter(id => id !== followerId)
        : [...prev.selectedFollowers, followerId];
      return { ...prev, selectedFollowers: selected };
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    // Validate private post has selected followers
    if (formData.privacy === 'private' && formData.selectedFollowers.length === 0) {
      setError('Please select at least one follower for private posts');
      setLoading(false);
      return;
    }

    try {
      const formDataToSend = new FormData();
      formDataToSend.append('title', formData.title);
      formDataToSend.append('content', formData.content);
      formDataToSend.append('category', formData.category);
      
      // Convert privacy to numeric level
      let privacyLevel = 0;
      if (formData.privacy === 'almost-private') privacyLevel = 1;
      if (formData.privacy === 'private') privacyLevel = 2;
      formDataToSend.append('privacy_level', privacyLevel);

      // Add selected followers for private posts
      if (formData.privacy === 'private') {
        formData.selectedFollowers.forEach(id => {
          formDataToSend.append('selected_followers', id);
        });
      }

      if (formData.image) {
        formDataToSend.append('imgOrgif', formData.image);
      }

      const response = await fetch('/api/create-post', {
        method: 'POST',
        body: formDataToSend,
        credentials: 'include'
      });

      if (response.ok) {
        router.push('/posts?message=Post created successfully');
      } else {
        const data = await response.json();
        setError(data.message || 'Failed to create post');
      }
    } catch (err) {
      setError('An error occurred. Please try again.');
      console.error('Create post error:', err);
    } finally {
      setLoading(false);
    }
  };

  // Listen for ReplaceButton file selection (custom event)
  useEffect(() => {
    const onReplace = (ev) => {
      const file = ev.detail;
      if (file) setFormData(prev => ({ ...prev, image: file }));
    };
    window.addEventListener('replace-image-selected', onReplace);
    return () => window.removeEventListener('replace-image-selected', onReplace);
  }, []);

  if (sessionLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="text-lg">Loading...</div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <Head>
        <title>Create Post - Share Your Story</title>
        <meta name="description" content="Create amazing content and inspire your community!" />
      </Head>
      
      <div className="create-post-page posts-page-bg">
        <div className="container-professional">
          
          {/* Professional Header */}
          <div className="create-post-hero">
            {/* <div className="create-post-hero-icon">
              <svg width="44" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            </div> */}
            <h1 className="create-post-hero-title">Create Something Amazing</h1> 
            <button
              onClick={() => router.back()}
              className="back-button"
            >
              ← Back to Posts
            </button>
          </div>
          {/* Ultra-Professional Form Container */}
          <div className="create-post-form-container">
            
            <div className="p-12">
              <div className="text-center mb-10">
                <h2 className="text-3xl font-bold mb-4" style={{color: '#4b5563'}}>
                  Share Your Story
                </h2>
               
              </div>

              <form onSubmit={handleSubmit}>
                <div className="space-y-8">
                  
                  {/* Title Field - Professional Design */}
                  <div className="space-y-4">
                    <label htmlFor="title" className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                      Post Title <span className="text-red-500">*</span>
                    </label>
                    <div className="relative group">
                      <div className="absolute inset-0 bg-gradient-to-r from-purple-400 to-pink-400 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-all duration-500"></div>
                      <input
                        id="title"
                        name="title"
                        type="text"
                        value={formData.title}
                        onChange={handleChange}
                        required
                        className="relative w-full px-8 py-6 border-3 border-gray-300 rounded-2xl bg-white font-bold text-xl transition-all duration-300 focus:border-purple-500 focus:ring-4 focus:ring-purple-100 focus:shadow-2xl outline-none hover:border-purple-400 hover:shadow-lg"
                        placeholder="Give your story an amazing title..."
                        style={{ borderWidth: '3px' }}
                      />
                    </div>
                   
                  </div>

                  {/* Content Field - Enhanced Textarea */}
                  <div className="space-y-4">
                    <label htmlFor="content" className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                      Post Content <span className="text-red-500">*</span>
                    </label>
                    <div className="relative group">
                      <div className="absolute inset-0 bg-gradient-to-r from-blue-400 to-cyan-400 rounded-2xl opacity-0 group-hover:opacity-20 blur-xl transition-all duration-500"></div>
                      <textarea
                        id="content"
                        name="content"
                        rows="8"
                        value={formData.content}
                        onChange={handleChange}
                        required
                        className="relative w-full px-8 py-6 border-3 border-gray-300 rounded-2xl bg-white font-semibold text-lg transition-all duration-300 focus:border-blue-500 focus:ring-4 focus:ring-blue-100 focus:shadow-2xl outline-none hover:border-blue-400 hover:shadow-lg resize-none"
                        placeholder="What's on your mind? Share your inspiration, thoughts, dreams..."
                        style={{ borderWidth: '3px' }}
                      />
                    </div>
                  </div>

                  {/* Category Field - Beautiful Selector */}
                  <div className="space-y-4">
                    <label htmlFor="category" className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                      Category <span className="text-red-500">*</span>
                    </label>
                    <div className="relative group">
                      <select
                        id="category"
                        name="category"
                        value={formData.category}
                        onChange={handleChange}
                        required
                        className="relative w-full px-8 py-6 border-3 rounded-2xl bg-white font-bold text-lg transition-all duration-300 outline-none hover:border-gray-400 hover:shadow-lg cursor-pointer"
                        style={{ borderWidth: '3px', borderColor: '#d1d5db', color: '#374151' }}
                      >
                        <option value="">Select a category...</option>
                        <option value="Travel">Travel</option>
                        <option value="Sport">Sports</option>
                        <option value="Food">Food</option>
                        <option value="Nature">Nature</option>
                      </select>
                    </div>
                  </div>

                  {/* Privacy Settings - Ultra-Modern */}
                  <div className="space-y-6">
                    <label className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                      Privacy Setting <span className="text-red-500">*</span>
                    </label>
                    
                    <div className="grid gap-5">
                      {privacyOptions.map((option) => {
                        const isSelected = formData.privacy === option.value;
                        const gradients = {
                          'public': 'linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%)',
                          'almost-private': 'linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%)',
                          'private': 'linear-gradient(135deg, #faf5ff 0%, #f3e8ff 100%)'
                        };
                        
                        return (
                          <label 
                            key={option.value} 
                            className={`relative flex items-center p-6 rounded-2xl cursor-pointer transition-all duration-300 border-3 backdrop-blur-sm group ${
                              isSelected 
                                ? 'border-transparent shadow-2xl scale-105 ring-4' 
                                : 'border-gray-300 bg-white hover:bg-gray-50 hover:border-gray-400 hover:shadow-lg hover:scale-[1.02]'
                            }`}
                            style={{
                              background: isSelected ? gradients[option.value] : 'white',
                              borderWidth: '3px',
                              ringColor: isSelected ? (option.value === 'public' ? '#10b98120' : option.value === 'almost-private' ? '#3b82f620' : '#f59e0b20') : 'transparent'
                            }}
                          >
                            <input
                              type="radio"
                              id={option.value}
                              name="privacy"
                              value={option.value}
                              checked={isSelected}
                              onChange={handleChange}
                              className="sr-only"
                            />
                            
                            <div className="flex items-center gap-6 relative z-10 w-full">
                              <div className="flex-1">
                                <div className={`text-xl font-black mb-2`} style={{
                                  color: isSelected ? '#667eea' : '#374151'
                                }}>
                                  {option.label}
                                </div>
                                <div className={`text-base font-semibold`} style={{
                                  color: isSelected ? '#667eea' : '#6b7280'
                                }}>
                                  {option.description}
                                </div>
                              </div>
                            </div>
                          </label>
                        );
                      })}
                    </div>
                    
                 
                  </div>

            {/* Follower Selector for Private Posts */}
            {formData.privacy === 'private' && (
              <div className="space-y-4">
                <label className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                  
                  Select Followers <span className="text-red-500">*</span>
                </label>
                <div className="border-3 border-indigo-300 rounded-2xl p-6 max-h-80 overflow-y-auto bg-gradient-to-br from-indigo-50 to-purple-50" style={{ borderWidth: '3px' }}>
                  {followers.length === 0 ? (
                    <div className="text-center py-8">
                      <p className="text-lg font-bold mb-2" style={{color: '#4b5563'}}>
                        No followers yet!
                      </p>
                      <p className="text-sm font-medium" style={{color: '#6b7280'}}>
                        Set post to "Almost Private" to allow future followers to see it.
                      </p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {followers.map((follower) => (
                        <label 
                          key={follower.id} 
                          className={`flex items-center p-4 rounded-xl cursor-pointer transition-all duration-300 border-2 ${
                            formData.selectedFollowers.includes(follower.id)
                              ? 'bg-white border-indigo-500 shadow-lg scale-[1.02]'
                              : 'bg-white/60 border-gray-200 hover:bg-white hover:border-indigo-300 hover:shadow-md'
                          }`}
                        >
                          <input
                            type="checkbox"
                            checked={formData.selectedFollowers.includes(follower.id)}
                            onChange={() => handleFollowerToggle(follower.id)}
                            className="w-5 h-5 mr-4 text-indigo-600 focus:ring-indigo-500 rounded border-2"
                          />
                          <div className="flex items-center gap-3 flex-1">
                            <span className="text-base font-bold" style={{color: '#4b5563'}}>{follower.username ? `@${follower.username}` : ''}</span>
                          </div>
                          {formData.selectedFollowers.includes(follower.id) 
                          }
                        </label>
                      ))}
                    </div>
                  )}
                </div>
                {formData.selectedFollowers.length > 0 && (
                  <div className="bg-green-50 border-2 border-green-300 rounded-xl p-4 flex items-center gap-3">
                    {/* <span className="text-2xl">✅</span> */}
                    <p className="text-base font-bold text-green-800">
                      {formData.selectedFollowers.length} follower(s) selected
                    </p>
                  </div>
                )}
              </div>
            )}

                  {/* Modern Compact File Upload */}
                  <div className="space-y-4">
                    <label className="block text-xl font-bold flex items-center gap-3" style={{color: '#4b5563'}}>
                      <span className="inline-flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-rose-500 text-white text-lg font-bold shadow-lg">
                        
                      </span>
                      Add Image <span style={{color: '#9ca3af', fontWeight: 'normal', fontSize: '0.875rem'}}>(Optional)</span>
                    </label>
                    
                    {!formData.image ? (
                      /* Upload Button */
                      <div className="relative">
                        <label 
                          htmlFor="image"
                          className="group flex items-center gap-4 p-6 rounded-2xl border-2 border-dashed cursor-pointer transition-all duration-300 hover:shadow-lg"
                          style={{ background: 'var(--primary-50)', borderColor: 'var(--primary-200)' }}
                        >
                          <div className="flex-shrink-0">
                            <div className="w-16 h-16 rounded-2xl flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform duration-300" style={{ background: 'linear-gradient(135deg, var(--primary-500), var(--primary-600))' }}>
                              <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 4v16m8-8H4" />
                              </svg>
                            </div>
                          </div>
                          
                          <div className="flex-1">
                            <h4 className="text-lg font-bold mb-1 transition-colors" style={{ color: '#4b5563' }}>
                              Click to upload an image
                            </h4>
                            <p className="text-sm font-medium" style={{color: '#6b7280'}}>
                              Supports: JPEG, PNG, GIF (Max 10MB)
                            </p>
                          </div>
                          
                          <div className="flex-shrink-0 flex items-center gap-3">
                            <svg className="w-6 h-6" style={{color: '#667eea'}} viewBox="0 0 24 24" fill="none" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                            </svg>
                          </div>
                        </label>
                        
                        <input
                          type="file"
                          id="image"
                          name="image"
                          accept="image/jpeg,image/jpg,image/png,image/gif"
                          onChange={handleChange}
                          className="sr-only"
                        />
                      </div>
                    ) : (
                      /* Image Preview Card */
                      <div className="bg-gradient-to-br from-green-50 to-emerald-50 rounded-2xl p-5 border-2 border-green-300 shadow-lg relative overflow-hidden">
                        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-green-400 to-emerald-500"></div>
                        
                        <div className="flex items-start gap-4">
                          {/* Image Thumbnail */}
                          <div className="flex-shrink-0 relative group">
                            <div className="w-24 h-24 rounded-xl overflow-hidden shadow-md border-2 border-white">
                              <img 
                                src={URL.createObjectURL(formData.image)} 
                                alt="Preview" 
                                className="w-full h-full object-cover"
                              />
                            </div>
                            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 rounded-xl transition-colors duration-200"></div>
                          </div>
                          
                          {/* File Info */}
                          <div className="flex-1 min-w-0">
                            <div className="flex items-start justify-between gap-2 mb-2">
                              <div className="flex items-center gap-2">
                                <svg className="w-5 h-5 text-green-600 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                                </svg>
                              </div>
                            </div>
                            
                            <p className="text-sm text-green-700 font-medium mb-3 truncate">
                              {formData.image.name}
                            </p>
                            
                            <div className="flex gap-2 image-action-group">
                              <button
                                type="button"
                                onClick={() => setFormData({...formData, image: null})}
                                className="inline-flex items-center gap-2 px-4 py-2 bg-white border-2 border-red-300 text-red-700 rounded-xl font-semibold text-sm hover:bg-red-50 hover:border-red-400 transition-all duration-200 shadow-sm hover:shadow image-action-btn image-action-remove"
                              >
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                                Remove
                              </button>
                              
                              {/* Replace button triggers hidden file input via ref for reliability */}
                              <ReplaceButton />
                            </div>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Professional Error Display */}
                  {error && (
                    <div className="bg-red-50/90 backdrop-blur-lg border-2 border-red-200 rounded-2xl p-6 relative overflow-hidden">
                      <div className="absolute top-0 left-0 right-0 h-1 bg-red-500"></div>
                      <div className="flex items-center gap-4">
                        <div className="w-6 h-6 bg-red-500 rounded-full flex items-center justify-center text-white font-bold">!</div>
                        <div>
                          <h4 className="text-lg font-bold text-red-800 mb-1">Error</h4>
                          <p className="text-red-700 font-semibold">{error}</p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Themed Submit Buttons */}
                  <div className="flex flex-row gap-6 pt-8">
                    {/* Create Post Button - use group share styles */}
                    <button
                      type="submit"
                      disabled={loading}
                      className="btn btn-primary flex-1"
                      style={{ minWidth: '200px' }}
                    >
                      {loading ? (
                        <>
                          <div className="w-6 h-6 border-3 border-white border-b-transparent rounded-full animate-spin mr-3" />
                          <span className="text-lg tracking-wider">Creating...</span>
                        </>
                      ) : (
                        <>
                          <span className="tracking-wider">Create Post</span>
                        </>
                      )}
                    </button>

                    {/* Cancel Button - use group share secondary style */}
                    <button
                      type="button"
                      onClick={() => router.push('/posts')}
                      className="btn btn-secondary"
                    >
                      <span className="tracking-wider">Cancel</span>
                     
                    </button>
                  </div>
                  
                </div>
              </form>
              
              {/* Inspirational Footer */}
              <div className="text-center mt-12 p-6 bg-gradient-to-r from-blue-50/50 via-purple-50/50 to-pink-50/50 rounded-2xl">
                <p className="text-lg font-medium text-gray-600">
                  Your story has the power to inspire the world!
                </p>
              </div>
            </div>
          </div>
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
        id="image-replace"
        name="image"
        accept="image/jpeg,image/jpg,image/png,image/gif"
        onChange={(e) => {
          /* propagate change to parent handler: find form input by name and trigger synthetic event */
          if (e.target.files && e.target.files[0]) {
            const file = e.target.files[0];
            /* find the global handler on page by dispatching a custom event */
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
