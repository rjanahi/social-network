import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Head from 'next/head';
import Layout from '../components/Layout';
import Toast from '../components/Toast';
import { useSession } from '../hooks/useSession';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export default function Groups() {
  const { user, loading } = useSession();
  const router = useRouter();
  const { subscribe, setNotifications } = useWebSocketContext();
  const [groups, setGroups] = useState([]);
  const [userGroups, setUserGroups] = useState([]);
  const [invitations, setInvitations] = useState([]);
  const [pendingApprovals, setPendingApprovals] = useState([]);
  const [loadingGroups, setLoadingGroups] = useState(true);
  const [activeTab, setActiveTab] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newGroup, setNewGroup] = useState({ title: '', description: '' });
  const [toast, setToast] = useState({ visible: false, message: '', type: 'info' });
  const [leaveConfirmId, setLeaveConfirmId] = useState(null);

  const showToast = (message, type = 'info', duration = 3500) => {
    setToast({ visible: true, message, type, duration });
  };

  const hideToast = () => setToast(prev => ({ ...prev, visible: false }));

  useEffect(() => {
    if (user && user.userID) {
      fetchGroups();
      fetchUserGroups();
      fetchInvitations();
      fetchPendingApprovals();
    }
  }, [user?.userID]);

  // Listen for real-time group updates
  useEffect(() => {
    if (!user) return;

    const unsubscribe = subscribe((message) => {
      console.log('[Groups] Received WebSocket message:', message);
      
      // Refresh group lists when relevant events occur
      if (message.type === 'group_invitation' || 
          message.type === 'group_join_request' ||
          message.type === 'join_request_sent' ||
          message.type === 'new_group_created' ||
          message.type === 'group_request' ||
          message.type === 'group_member_update' ||
          message.type === 'group_member_left' ||
          message.type === 'invitation_response' ||
          message.type === 'group_request_response') {
        console.log('[Groups] Refreshing group data...');
        fetchGroups();
        fetchUserGroups();
        fetchInvitations();
        fetchPendingApprovals();
      }
    });

    return () => unsubscribe();
  }, [user, subscribe]);

  const fetchGroups = async () => {
    try {
      const response = await fetch('/api/groups', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        const groupsList = data?.groups || [];
        setGroups(groupsList);
      } else {
        console.error('Failed to fetch groups:', response.status);
        setGroups([]);
      }
    } catch (error) {
      console.error('Error fetching groups:', error);
      setGroups([]);
    }
    setLoadingGroups(false);
  };

  const fetchUserGroups = async () => {
    try {
      const response = await fetch('/api/user-groups', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        setUserGroups(data?.groups || []);
      } else {
        console.error('Failed to fetch user groups:', response.status);
        setUserGroups([]);
      }
    } catch (error) {
      console.error('Error fetching user groups:', error);
      setUserGroups([]);
    }
  };

  const fetchInvitations = async () => {
    try {
      const response = await fetch('/api/groups/invitations', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        console.log('[fetchInvitations] Received data:', data);
        console.log('[fetchInvitations] Invitations array:', data?.invitations);
        setInvitations(data?.invitations || []);
      } else {
        console.error('Failed to fetch invitations:', response.status);
        setInvitations([]);
      }
    } catch (error) {
      console.error('Error fetching invitations:', error);
      setInvitations([]);
    }
  };

  const fetchPendingApprovals = async () => {
    try {
      const response = await fetch('/api/groups/pending-approvals', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        setPendingApprovals(data?.requests || []);
      } else {
        console.error('Failed to fetch pending approvals:', response.status);
        setPendingApprovals([]);
      }
    } catch (error) {
      console.error('Error fetching pending approvals:', error);
      setPendingApprovals([]);
    }
  };

  const handleCreateGroup = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/create-group', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newGroup),
        credentials: 'include'
      });
      
      if (response.ok) {
        setShowCreateModal(false);
        setNewGroup({ title: '', description: '' });
        fetchGroups();
      } else {
        console.error('Failed to create group:', response.status);
      }
    } catch (error) {
      console.error('Error creating group:', error);
    }
  };

  const handleJoinGroup = async (groupId) => {
    try {
      const response = await fetch('/api/join-group', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ group_id: groupId }),
        credentials: 'include'
      });
      
      if (response.ok) {
        // Refresh data to get updated membership status
        fetchGroups();
      } else {
        const errorData = await response.json();
        console.error('Failed to join group:', response.status, errorData);
        if (response.status === 409) {
          showToast('You have already requested to join this group or are already a member.', 'info');
        } else {
          showToast(`Failed to join group: ${errorData.error || 'Unknown error'}`, 'error');
        }
      }
    } catch (error) {
      console.error('Error joining group:', error);
    }
  };

  const confirmLeaveGroup = async (groupId) => {
    try {
      const response = await fetch('/api/leave-group', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ group_id: groupId }),
        credentials: 'include'
      });

      if (response.ok) {
        showToast('Successfully left the group', 'success');
        // Refresh all group data
        fetchGroups();
        fetchUserGroups();
      } else {
        const data = await response.json();
        showToast(data.error || 'Failed to leave group', 'error');
        console.error('Failed to leave group:', response.status);
      }
    } catch (error) {
      console.error('Error leaving group:', error);
      showToast('Error leaving group', 'error');
    } finally {
      setLeaveConfirmId(null);
    }
  };

  const handleApproveRequest = async (requestId, action) => {
    try {
      // Optimistic removal of matching notifications (group_join_request) so bell updates immediately
      try {
        setNotifications(prev => {
          if (!prev) return prev;
          return prev.filter(n => {
            if (!n) return true;
            // If this notification is a join request and references this requestId, remove it
            const rid = n.RequestID || n.request_id || n.requestId || n.id || 0;
            if (rid && Number(rid) === Number(requestId)) return false;

            // Also remove notifications that reference the same group id (fallback)
            const ngid = n.GroupID || n.group_id || n.groupId || n.group || null;
            if (ngid && String(ngid) === String(requestId)) return false;

            return true;
          });
        });
      } catch (e) {
        console.error('Optimistic notification removal failed for approveRequest', e);
      }

      const response = await fetch('/api/groups/respond-request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          request_id: requestId, 
          action: action // 'accept' or 'decline'
        }),
        credentials: 'include'
      });

      if (response.ok) {
        // Refresh pending approvals
        fetchPendingApprovals();
        showToast(`Request ${action}ed successfully!`, 'success');
        // run a delayed cleanup in case websocket re-broadcasts the same notification
        setTimeout(() => {
          setNotifications(prev => {
            if (!prev) return prev;
            return prev.filter(n => {
              const rid = n.RequestID || n.request_id || n.requestId || n.id || 0;
              const ngid = n.GroupID || n.group_id || n.groupId || n.group || null;
              if (rid && Number(rid) === Number(requestId)) return false;
              if (ngid && String(ngid) === String(requestId)) return false;
              return true;
            });
          });
        }, 300);
      } else {
        const txt = await response.text();
        console.error(`Failed to ${action} request:`, response.status, txt);
        // restore pending approvals from server
        fetchPendingApprovals();
        showToast(`Failed to ${action} request: ${txt || response.status}`, 'error');
      }
    } catch (error) {
      console.error(`Error ${action}ing request:`, error);
      showToast(`Error ${action}ing request`, 'error');
    }
  };

  const handleRespondToInvitation = async (invitationId, action) => {
    try {
      // Find the related invitation object (if available) so we can match notifications by group+inviter
      const relatedInv = (invitations || []).find(inv => {
        return String(inv.id) === String(invitationId)
          || String(inv.group_id) === String(invitationId)
          || String(inv.groupId) === String(invitationId)
          || (inv && inv.id && String(inv.id) === String(invitationId));
      }) || null;

      // Optimistically remove matching notifications and invitation from UI so user sees immediate feedback.
      try {
        setNotifications(prev => {
          if (!prev) return prev;
          return prev.filter(n => {
            if (!n) return true;
            // remove common invitation/join request types that refer to this group/invitation
            const nid = n.InvitationID || n.invitation_id || n.invitationId || n.id || 0;
            if (nid && relatedInv && Number(nid) === Number(relatedInv.id)) return false;

            const ngid = n.GroupID || n.group_id || n.groupId || n.group || null;
            const ninv = n.From || n.from || n.inviter_id || n.inviterId || null;

            // If relatedInv exists, match by group + inviter (robust to missing invitation id)
            if (relatedInv && (ngid && String(ngid) === String(relatedInv.group_id))) return false;
            if (relatedInv && ninv && (String(ninv) === String(relatedInv.inviter_id) || String(ninv) === String(relatedInv.inviter))) return false;

            // Fallback: if the notification directly references the group id passed in
            if (ngid && String(ngid) === String(invitationId)) return false;

            return true;
          });
        });
        // Also remove the invitation from the invitations list optimistically
        setInvitations(prev => (prev || []).filter(inv => String(inv.id) !== String(invitationId) && String(inv.group_id) !== String(invitationId)));

        // After a short delay, run one more cleanup pass in case the websocket re-broadcasted a related notification
        setTimeout(() => {
          setNotifications(prev => {
            if (!prev) return prev;
            return prev.filter(n => {
              if (!n) return true;
              const ngid = n.GroupID || n.group_id || n.groupId || n.group || null;
              const nid = n.InvitationID || n.invitation_id || n.invitationId || n.id || 0;
              if (relatedInv && ((nid && Number(nid) === Number(relatedInv.id)) || (ngid && String(ngid) === String(relatedInv.group_id)))) return false;
              if (ngid && String(ngid) === String(invitationId)) return false;
              return true;
            });
          });
        }, 300);
      } catch (e) {
        console.error('Optimistic removal failed', e);
      }

      const response = await fetch('/api/groups/respond-invitation', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          invitation_id: invitationId, 
          action: action // 'accept' or 'decline'
        }),
        credentials: 'include'
      });
      
      if (response.ok) {
        // Refresh invitations and groups to ensure canonical state
        fetchInvitations();
        fetchGroups();
        fetchUserGroups();
        showToast(`Invitation ${action}ed successfully!`, 'success');
        // already removed optimistically above
      } else {
        const txt = await response.text();
        console.error(`Failed to ${action} invitation:`, response.status, txt);
        // Restore invitations and notifications by re-fetching server state
        fetchInvitations();
        showToast(`Failed to ${action} invitation: ${txt || response.status}`, 'error');
      }
    } catch (error) {
      console.error(`Error ${action}ing invitation:`, error);
      showToast(`Error ${action}ing invitation`, 'error');
    }
  };

  const viewGroupDetails = (groupId) => {
    router.push(`/groupDetails?id=${groupId}`);
  };

  const getCurrentGroups = () => {
    switch (activeTab) {
      case 'all':
        return groups;
      case 'my':
        return userGroups;
      case 'invitations':
        return invitations;
      case 'pending':
        return pendingApprovals;
      default:
        return groups;
    }
  };

  if (loading || loadingGroups) {
    return (
      <Layout>
        <div className="loading-container">
          <div className="loading-text">Loading groups...</div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <Head>
        <title>Groups - Join Amazing Communities</title>
        <meta name="description" content="Find your perfect tribe and connect with incredible people!" />
      </Head>
      
      <div className="groups-page-modern">
        {/* Hero Section */}
        <div className="groups-hero-section">
          <div className="groups-hero-content">
            <div className="groups-hero-icon">
              <svg width="64" height="64" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="url(#gradient1)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <circle cx="9" cy="7" r="4" stroke="url(#gradient1)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="url(#gradient2)" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <defs>
                  <linearGradient id="gradient1" x1="0" y1="0" x2="24" y2="24">
                    <stop offset="0%" stopColor="#667eea"/>
                    <stop offset="100%" stopColor="#764ba2"/>
                  </linearGradient>
                  <linearGradient id="gradient2" x1="0" y1="0" x2="24" y2="24">
                    <stop offset="0%" stopColor="#f093fb"/>
                    <stop offset="100%" stopColor="#f5576c"/>
                  </linearGradient>
                </defs>
              </svg>
            </div>
            
            <h1 className="groups-hero-title">
              Discover Your Community
            </h1>
            
            <p className="groups-hero-subtitle">
              Join vibrant groups, share experiences, and connect with like-minded people
            </p>

            <button 
              onClick={() => setShowCreateModal(true)}
              className="groups-hero-cta"
            >
              <span className="groups-hero-cta-icon">+</span>
              Create New Group
            </button>
          </div>
        </div>
        
        <div className="groups-section container-professional">
          <div className="groups-container-modern">

          {/* Modern Create Group Modal */}
          {showCreateModal && (
            <div className="modal-overlay-modern" onClick={() => setShowCreateModal(false)}>
              <div className="modal-content-modern" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header-modern">
                  <div className="modal-header-icon">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2"/>
                    </svg>
                  </div>
                  <h2 className="modal-title-modern">Create New Group</h2>
                  <button className="modal-close-modern" onClick={() => setShowCreateModal(false)}>
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  </button>
                </div>
                
                <form onSubmit={handleCreateGroup} className="modal-form-modern">
                  <div className="form-group-modern">
                    <label className="form-label-modern">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M4 6h16M4 12h16M4 18h16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                      Group Name
                    </label>
                    <input
                      type="text"
                      className="form-input-modern"
                      placeholder="Enter a catchy group name..."
                      value={newGroup.title}
                      onChange={(e) => setNewGroup({...newGroup, title: e.target.value})}
                      required
                    />
                  </div>
                  
                  <div className="form-group-modern">
                    <label className="form-label-modern">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8l-6-6z" stroke="currentColor" strokeWidth="2"/>
                        <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                      Description
                    </label>
                    <textarea
                      className="form-textarea-modern"
                      placeholder="Describe what your group is about..."
                      value={newGroup.description}
                      onChange={(e) => setNewGroup({...newGroup, description: e.target.value})}
                      rows="4"
                      required
                    />
                  </div>
                  
                  <div className="modal-actions-modern">
                    <button 
                      type="button" 
                      onClick={() => setShowCreateModal(false)} 
                      className="modal-btn-modern modal-btn-cancel"
                    >
                      Cancel
                    </button>
                    <button 
                      type="submit" 
                      className="modal-btn-modern modal-btn-create"
                    >
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                      Create Group
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          {/* Modern Tabs */}
          <div className="groups-tabs-modern">
            <button
              onClick={() => setActiveTab('all')}
              className={`groups-tab-modern ${activeTab === 'all' ? 'active' : ''}`}
            >
              <span className="groups-tab-icon">🌐</span>
              <span className="groups-tab-label">All Groups</span>
              <span className="groups-tab-count">{groups.length}</span>
            </button>
            
            <button
              onClick={() => setActiveTab('my')}
              className={`groups-tab-modern ${activeTab === 'my' ? 'active' : ''}`}
            >
              <span className="groups-tab-icon">👥</span>
              <span className="groups-tab-label">My Groups</span>
              <span className="groups-tab-count">{userGroups.length}</span>
            </button>
            
            <button
              onClick={() => setActiveTab('invitations')}
              className={`groups-tab-modern ${activeTab === 'invitations' ? 'active' : ''}`}
            >
              <span className="groups-tab-icon">✉️</span>
              <span className="groups-tab-label">Invitations</span>
              {invitations.length > 0 && (
                <span className="groups-tab-badge">{invitations.length}</span>
              )}
            </button>
            
            <button
              onClick={() => setActiveTab('pending')}
              className={`groups-tab-modern ${activeTab === 'pending' ? 'active' : ''}`}
            >
              <span className="groups-tab-icon">⏳</span>
              <span className="groups-tab-label">Pending</span>
              {pendingApprovals.length > 0 && (
                <span className="groups-tab-badge">{pendingApprovals.length}</span>
              )}
            </button>
          </div>

          {/* Groups Grid */}
          <div className="groups-grid-modern">
            {getCurrentGroups().map((group) => {
              if (activeTab === 'invitations') {
                console.log('[Render] Invitation group:', group);
              }
              return (
              <div key={group.id} className="group-card-modern">
                {/* Card Header with Gradient */}
                <div className="group-card-header-modern">
                  <div className="group-card-icon-wrapper">
                    <div className="group-card-icon">
                      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    </div>
                  </div>
                  {group.member_count !== undefined && (
                    <div className="group-card-members-badge">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2"/>
                        <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                      </svg>
                      <span>{group.member_count || 0}</span>
                    </div>
                  )}
                </div>
                
                {/* Card Body */}
                <div className="group-card-body-modern">
                  <h3 className="group-card-title-modern">{group.title || group.group_name}</h3>
                  <p className="group-card-description-modern">{group.description}</p>
                  
                  <div className="group-card-meta-modern">
                    <div className="group-card-creator">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        <circle cx="12" cy="7" r="4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                      <span>
                        {activeTab === 'invitations' 
                          ? `${group.inviter_name || group.inviter}` 
                          : `${group.creator_name || group.creator || group.username}`}
                      </span>
                    </div>
                    {group.created_at && (
                      <div className="group-card-date">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                          <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                        </svg>
                        <span>{new Date(group.created_at).toLocaleDateString()}</span>
                      </div>
                    )}
                  </div>
                </div>                {/* Card Actions */}
                <div className="group-card-actions-modern">
                  {(() => {
                    // Handle different tab types
                    if (activeTab === 'pending') {
                      // For pending approvals tab - show approve/decline buttons
                      return (
                        <>
                          <button
                            onClick={() => handleApproveRequest(group.id, 'accept')}
                            className="group-btn-modern group-btn-success"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M20 6L9 17l-5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Approve
                          </button>
                          <button
                            onClick={() => handleApproveRequest(group.id, 'decline')}
                            className="group-btn-modern group-btn-danger"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Decline
                          </button>
                        </>
                      );
                    }

                    if (activeTab === 'invitations') {
                      return (
                        <>
                          <button
                            onClick={() => handleRespondToInvitation(group.id, 'accept')}
                            className="group-btn-modern group-btn-success"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M20 6L9 17l-5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Accept
                          </button>
                          <button
                            onClick={() => handleRespondToInvitation(group.id, 'decline')}
                            className="group-btn-modern group-btn-secondary"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Decline
                          </button>
                        </>
                      );
                    }

                    // Regular group actions for 'all' and 'my' tabs
                    const isCreator = user?.userID === group.creator_id;
                    const membershipStatus = group.user_membership_status;
                    const isMember = activeTab === 'my' || membershipStatus === 'accepted';
                    const isPending = membershipStatus === 'pending';

                    if (isCreator) {
                      return (
                        <>
                          <button
                            onClick={() => viewGroupDetails(group.id)}
                            className="group-btn-modern group-btn-primary"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Enter Group
                          </button>
                        </>
                      );
                    } else if (isMember) {
                      return (
                        <>
                          <button
                            onClick={() => viewGroupDetails(group.id)}
                            className="group-btn-modern group-btn-primary"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Enter Group
                          </button>
                          {leaveConfirmId === group.id ? (
                            <>
                              <button
                                onClick={() => confirmLeaveGroup(group.id)}
                                className="group-btn-modern group-btn-danger-outline"
                              >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                  <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                                </svg>
                                Confirm
                              </button>
                              <button
                                onClick={() => setLeaveConfirmId(null)}
                                className="group-btn-modern group-btn-secondary"
                              >
                                Cancel
                              </button>
                            </>
                          ) : (
                            <button
                              onClick={() => setLeaveConfirmId(group.id)}
                              className="group-btn-modern group-btn-danger-outline"
                            >
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4M16 17l5-5-5-5M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                              </svg>
                              Leave
                            </button>
                          )}
                        </>
                      );
                    } else if (isPending) {
                      return (
                        <>
                          <button disabled className="group-btn-modern group-btn-pending">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                              <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                            </svg>
                            Pending
                          </button>
                          <button
                            onClick={() => viewGroupDetails(group.id)}
                            className="group-btn-modern group-btn-secondary"
                          >
                            View
                          </button>
                        </>
                      );
                    } else {
                      return (
                        <>
                          <button
                            onClick={() => handleJoinGroup(group.id)}
                            className="group-btn-modern group-btn-success"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M12.5 7a4 4 0 11-8 0 4 4 0 018 0zM16 11h6M19 8v6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Join
                          </button>
                          <button
                            onClick={() => viewGroupDetails(group.id)}
                            className="group-btn-modern group-btn-secondary"
                          >
                            View
                          </button>
                        </>
                      );
                    }
                  })()}
                </div>
              </div>
              );
            })}
          </div>

            {getCurrentGroups().length === 0 && (
              <div className="no-groups-message-modern">
                <div className="no-groups-icon">
                  {activeTab === 'all' ? '🌐' : 
                   activeTab === 'my' ? '👥' : 
                   activeTab === 'invitations' ? '✉️' :
                   '⏳'}
                </div>
                <h3>
                  {activeTab === 'all' ? 'No groups available yet' : 
                   activeTab === 'my' ? 'You haven\'t joined any groups' : 
                   activeTab === 'invitations' ? 'No pending invitations' :
                   'No pending approvals'}
                </h3>
                <p>
                  {activeTab === 'all' ? 'Be the first to create a community!' : 
                   activeTab === 'my' ? 'Explore and join groups that interest you' : 
                   activeTab === 'invitations' ? 'Check back later for new invitations' :
                   'No requests waiting for your approval'}
                </p>
                {activeTab === 'all' && (
                  <button 
                    onClick={() => setShowCreateModal(true)}
                    className="no-groups-cta"
                  >
                    Create First Group
                  </button>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
      <Toast visible={toast.visible} message={toast.message} type={toast.type} onClose={hideToast} duration={toast.duration} />
    </Layout>
  );
}
