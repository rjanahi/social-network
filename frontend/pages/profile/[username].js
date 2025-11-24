import { useState, useEffect } from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import Layout from '../../components/Layout';
import { useSession } from '../../hooks/useSession';
import { useWebSocketContext } from '../../contexts/WebSocketContext';

export default function ProfilePage() {
  const router = useRouter();
  const { username } = router.query;
  const { user, loading: sessionLoading } = useSession();
  const { subscribe } = useWebSocketContext();

  const [profile, setProfile] = useState(null);
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  const fetchProfile = async () => {
    if (!username) return;
    setLoading(true);
    try {
      const resp = await fetch(`/api/get-otherPosts?username=${encodeURIComponent(username)}`, { credentials: 'include' });
      if (resp.ok) {
        const data = await resp.json();
        setProfile(data.profile || null);
        setPosts(data.posts || []);
      } else {
        console.error('Failed to fetch profile', resp.status);
        setProfile(null);
        setPosts([]);
      }
    } catch (err) {
      console.error('Error fetching profile', err);
      setProfile(null);
      setPosts([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!sessionLoading) fetchProfile();
  }, [username, sessionLoading]);

  // Listen for follow request responses: if someone accepted our follow request
  // while we're viewing their profile, refresh the profile automatically so it opens.
  useEffect(() => {
    if (!user) return;
    const unsubscribe = subscribe((message) => {
      try {
        const t = (message && (message.type || message.Type));
        if (t !== 'follow_request_response') return;

        // Ensure this message is addressed to the current user
        const to = message.to || message.To || message.ToUser || null;
        const from = message.from || message.From || null;
        const usernameFrom = message.username || message.Username || null;

        const currentUserId = user?.userID || user?.id;
        if (!currentUserId) return;

        if (parseInt(to) !== parseInt(currentUserId)) return;

        // If we're viewing this user's profile (by username) or the sender id matches
        if ((username && usernameFrom && usernameFrom === username) || (from && profile && (parseInt(from) === parseInt(profile.id)))) {
          // Refresh profile so it becomes viewable immediately
          fetchProfile();
        }
      } catch (err) {
        // ignore
      }
    });

    return () => unsubscribe && typeof unsubscribe === 'function' && unsubscribe();
  }, [user, subscribe, profile, username]);

  const togglePrivacy = async () => {
    setActionLoading(true);
    try {
      const resp = await fetch('/api/toggle-privacy', { method: 'POST', credentials: 'include' });
      if (resp.ok) {
        await fetchProfile();
      } else {
        console.error('Failed to toggle privacy');
      }
    } catch (e) {
      console.error('Toggle error', e);
    } finally {
      setActionLoading(false);
    }
  };

  const handleFollow = async () => {
    if (!profile) return;
    setActionLoading(true);
    try {
      const resp = await fetch('/api/follow-user', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username: profile.username })
      });
      if (resp.ok) await fetchProfile();
    } catch (e) {
      console.error('Follow error', e);
    } finally {
      setActionLoading(false);
    }
  };

  const handleUnfollow = async () => {
    if (!profile) return;
    setActionLoading(true);
    try {
      const resp = await fetch('/api/unfollow-user', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username: profile.username })
      });
      if (resp.ok) await fetchProfile();
    } catch (e) {
      console.error('Unfollow error', e);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <Layout>
        <div className="loading-container"><div className="loading-text">Loading profile...</div></div>
      </Layout>
    );
  }

  if (!profile) {
    return (
      <Layout>
        <div style={{ padding: '2rem' }}>
          <h2>Profile not found</h2>
        </div>
      </Layout>
    );
  }

  const isOwner = user && user.username && user.username === profile.username;
  const canView = profile.canView === true || isOwner || profile.isPrivate === false;

  return (
    <Layout>
      <Head>
        <title>{profile.fname} {profile.lname} - Profile</title>
      </Head>

      <div style={{ maxWidth: 1000, margin: '1.5rem auto' }}>
        <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center', marginBottom: '1rem' }}>
          <img src={profile.avatar || '/css/icon.png'} alt={profile.username} style={{ width:96,height:96,borderRadius:'50%' }} />
          <div style={{ flex: 1 }}>
            <h1 style={{ fontSize: '1.75rem' }}>{profile.fname} {profile.lname} {profile.nickname && <span style={{ fontSize: '1rem', color:'#6b7280' }}>"{profile.nickname}"</span>}</h1>
            <p style={{ color:'#6b7280' }}>@{profile.username}</p>
            <div style={{ display:'flex', gap: '1rem', marginTop: '0.5rem' }}>
              <div><strong>{profile.followerCount}</strong> Followers</div>
              <div><strong>{profile.followingCount}</strong> Following</div>
              <div><strong>{posts.length}</strong> Posts</div>
            </div>
          </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
            {isOwner ? (
              <>
                <button onClick={() => router.push('/myProfile')} className="icon-btn like-btn">Edit Profile</button>
                <button onClick={togglePrivacy} className={profile.isPrivate ? 'icon-btn dislike-btn' : 'icon-btn comment-btn'} disabled={actionLoading}>{profile.isPrivate ? 'Private' : 'Public'}</button>
              </>
            ) : (
              <>
                {/* Show follow/request/unfollow depending on profile state */}
                {profile.isFollowing ? (
                  <button onClick={handleUnfollow} className="icon-btn dislike-btn" disabled={actionLoading}>Unfollow</button>
                ) : (
                  profile.isPrivate ? (
                    // Private account: show Request or Pending
                    profile.requestStatus === 'pending' ? (
                      <button className="icon-btn comment-btn" disabled>Request Pending</button>
                    ) : (
                      <button onClick={handleFollow} className="icon-btn like-btn" disabled={actionLoading}>Request</button>
                    )
                  ) : (
                    // Public account: normal follow
                    <button onClick={handleFollow} className="icon-btn like-btn" disabled={actionLoading}>Follow</button>
                  )
                )}
              </>
            )}
          </div>
        </div>

        {(!canView) && (
          <div style={{ padding: '1rem', background:'#fffbeb', borderRadius:8, marginBottom: '1rem' }}>
            <strong>This is a private profile.</strong> Follow to see more details and posts.
          </div>
        )}

        {/* Profile info */}
        <div style={{ background: '#fff', padding: '1rem', borderRadius:8, marginBottom: '1rem' }}>
          <h3>About</h3>
          {canView ? (
            <>
              {profile.bio && <p>{profile.bio}</p>}
              <div>Email: {profile.email || '—'}</div>
              <div>Age: {profile.age || '—'}</div>
              <div>Gender: {profile.gender || '—'}</div>
              {profile.dateOfBirth && <div>Born: {new Date(profile.dateOfBirth).toLocaleDateString()}</div>}
            </>
          ) : (
            <p>Profile details are private.</p>
          )}
        </div>

        {/* Posts */}
        {canView && (
          <div style={{ marginBottom: '1rem' }}>
            <h3>Posts ({posts.length})</h3>
            {posts.length === 0 ? (
              <p>No posts yet.</p>
            ) : (
              <div style={{ display: 'grid', gap: '0.75rem' }}>
                {posts.map(p => (
                  <div key={p.id} style={{ padding: '1rem', background:'#fff', borderRadius:8, border: '1px solid #e5e7eb' }}>
                    <div style={{ display:'flex', justifyContent:'space-between', marginBottom:6 }}>
                      <strong>{p.title}</strong>
                      <span style={{ color:'#6b7280' }}>{p.createdAt}</span>
                    </div>
                    <div style={{ color:'#4b5563' }}>{p.content}</div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Followers / Following lists - only when allowed */}
        <div style={{ display:'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
          <div style={{ background:'#fff', padding: '1rem', borderRadius:8 }}>
            <h4>Followers ({profile.followerCount})</h4>
            {canView && profile.followers && profile.followers.length > 0 ? (
              <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
                {profile.followers.map(f => (
                  <div key={f.id} style={{ display:'flex', justifyContent:'space-between', alignItems:'center' }}>
                    <div><strong>{f.firstname} {f.lastname}</strong> <div style={{ color:'#6b7280' }}>@{f.username}</div></div>
                    <button onClick={() => router.push(`/profile/${encodeURIComponent(f.username)}`)} className="icon-btn comment-btn">View</button>
                  </div>
                ))}
              </div>
            ) : (
              <p style={{ color:'#6b7280' }}>Not available</p>
            )}
          </div>

          <div style={{ background:'#fff', padding: '1rem', borderRadius:8 }}>
            <h4>Following ({profile.followingCount})</h4>
            {canView && profile.following && profile.following.length > 0 ? (
              <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
                {profile.following.map(f => (
                  <div key={f.id} style={{ display:'flex', justifyContent:'space-between', alignItems:'center' }}>
                    <div><strong>{f.firstname} {f.lastname}</strong> <div style={{ color:'#6b7280' }}>@{f.username}</div></div>
                    <button onClick={() => router.push(`/profile/${encodeURIComponent(f.username)}`)} className="icon-btn comment-btn">View</button>
                  </div>
                ))}
              </div>
            ) : (
              <p style={{ color:'#6b7280' }}>Not available</p>
            )}
          </div>
        </div>
      </div>
    </Layout>
  );
}
