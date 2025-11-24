import { useState, useEffect } from 'react';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import Toast from './Toast';

// Use NEXT_PUBLIC_BACKEND_URL for client-side code; fallback to localhost for local dev
const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL || process.env.BACKEND_URL || 'http://localhost:8080';

export default function GroupEvents({ groupId, user }) {
  const [events, setEvents] = useState([]);
  const [loadingEvents, setLoadingEvents] = useState(true);
  const [showCreateEvent, setShowCreateEvent] = useState(false);
  const [newEvent, setNewEvent] = useState({ 
    title: '', 
    description: '', 
    date: '',
    time: ''
  });
  const { subscribe } = useWebSocketContext();
  const [toast, setToast] = useState({ visible: false, message: '', type: 'info' });

  const showToast = (message, type = 'info', duration = 3500) => {
    setToast({ visible: true, message, type, duration });
  };

  const hideToast = () => setToast(prev => ({ ...prev, visible: false }));

  useEffect(() => {
    if (groupId) {
      fetchGroupEvents();
    }
  }, [groupId]);

  // Listen for real-time event updates
  useEffect(() => {
    if (!user || !groupId) return;

    const unsubscribe = subscribe((message) => {
      console.log('[GroupEvents] Received WebSocket message:', message);
      if (message.group_id == groupId) {
        if (message.type === 'new_groupEvent') {
          // Add new event to state instantly using info from message
          setEvents(prev => {
            if (prev.some(ev => ev.id === message.post_id)) return prev;
            return [
              {
                id: message.post_id,
                title: message.content.split(':')[0]?.trim() || '',
                description: message.content.split(':').slice(1).join(':').trim() || '',
                created_at: message.timestamp,
                event_date: message.event_date || message.EventDate || null,
                yes_count: message.yes_count,
                no_count: message.no_count,
                user_vote: message.user_vote || null
              },
              ...prev
            ];
          });
        } else if (message.type === 'group_event_vote_update') {
          // Use info from WebSocket message to update event in state
          setEvents(prev => prev.map(ev =>
            ev.id === message.post_id
              ? {
                  ...ev,
                  yes_count: message.yes_count,
                  no_count: message.no_count,
                  user_vote: message.user_vote
                }
              : ev
          ));
        }
      }
    });
    return () => unsubscribe();
  }, [user, groupId, subscribe]);

  const fetchGroupEvents = async () => {
    if (!groupId) return;
    
    try {
      setLoadingEvents(true);
      console.log('Fetching group events for groupId:', groupId);
      
      const response = await fetch(`${BACKEND_URL}/groups/${groupId}/events`, {
        method: 'GET',
        credentials: 'include'
      });
      
      console.log('Events response status:', response.status);
      
      if (response.ok) {
        const eventsData = await response.json();
        console.log('Events data received:', eventsData);
        setEvents(Array.isArray(eventsData) ? eventsData : []);
      } else {
        const errorText = await response.text();
        console.error('Failed to fetch group events:', response.status, errorText);
        setEvents([]);
      }
    } catch (error) {
      console.error('Error fetching group events:', error);
      setEvents([]);
    } finally {
      setLoadingEvents(false);
    }
  };

  const handleCreateEvent = async (e) => {
    e.preventDefault();
    if (!newEvent.title.trim() || !newEvent.description.trim() || !newEvent.date || !newEvent.time) {
      showToast('Please fill in all fields', 'error');
      return;
    }
    
    try {
      const datetime = new Date(`${newEvent.date}T${newEvent.time}`);
      if (datetime < new Date()) {
        showToast('Event cannot be scheduled in the past', 'error');
        return;
      }

      const response = await fetch(`${BACKEND_URL}/create-event`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          group_id: parseInt(groupId),
          title: newEvent.title.trim(),
          description: newEvent.description.trim(),
          event_date: datetime.toISOString()
        }),
        credentials: 'include'
      });
      
      if (response.ok) {
        setShowCreateEvent(false);
        setNewEvent({ title: '', description: '', date: '', time: '' });
        fetchGroupEvents(); // Refresh events
      } else {
        const errorText = await response.text();
        showToast(`Failed to create event: ${errorText}`, 'error');
      }
    } catch (error) {
      console.error('Error creating event:', error);
      showToast('Error creating event', 'error');
    }
  };

  const handleVote = async (eventId, choice) => {
    try {
      const response = await fetch(`${BACKEND_URL}/groups/events/vote?event_id=${eventId}&choice=${choice}`, {
        method: 'POST',
        credentials: 'include'
      });
      // No reload! WebSocket will update the UI
      if (!response.ok) {
        console.error('Failed to vote on event');
      }
    } catch (error) {
      console.error('Error voting on event:', error);
    }
  };

  return (
    <div className="group-events-modern">
      {/* Modern Create Event Prompt */}
      <div className="create-event-prompt-modern">
        <button
          onClick={() => setShowCreateEvent(true)}
          className="create-event-button-modern"
        >
          <div className="create-event-icon">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
              <path d="M16 2v4M8 2v4M3 10h18M12 14h.01M12 18h.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
          <div className="create-event-text">
            <span className="create-event-title">Plan an Event</span>
            <span className="create-event-subtitle">Create something exciting for your group!</span>
          </div>
        </button>
      </div>

      {/* Create Event Modal */}
      {showCreateEvent && (
        <div className="modal">
          <div className="modal-content">
            <span className="close" onClick={() => setShowCreateEvent(false)}>&times;</span>
            <h2>Create New Group Event</h2>
            
            <form onSubmit={handleCreateEvent}>
              <div className="form-group">
                <label>Event Title</label>
                <input
                  type="text"
                  value={newEvent.title}
                  onChange={(e) => setNewEvent({...newEvent, title: e.target.value})}
                  placeholder="What's the event about?"
                  required
                />
              </div>
              
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={newEvent.description}
                  onChange={(e) => setNewEvent({...newEvent, description: e.target.value})}
                  rows="4"
                  placeholder="Describe the event details..."
                  required
                />
              </div>

              <div className="form-group">
                <label>Date</label>
                <input
                  type="date"
                  value={newEvent.date}
                  onChange={(e) => setNewEvent({...newEvent, date: e.target.value})}
                  required
                  min={new Date().toISOString().split('T')[0]}
                />
              </div>

              <div className="form-group">
                <label>Time</label>
                <input
                  type="time"
                  value={newEvent.time}
                  onChange={(e) => setNewEvent({...newEvent, time: e.target.value})}
                  required
                />
              </div>
              
              <div className="button-group-modern">
                <button type="button" onClick={() => setShowCreateEvent(false)} className="btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Create Event
                </button>
              </div>
            </form>
          </div>
            <Toast visible={toast.visible} message={toast.message} type={toast.type} onClose={hideToast} duration={toast.duration} />
        </div>
      )}

            {/* Events List */}
      {loadingEvents ? (
        <div className="loading-container">
          <div className="loading-text">Loading events...</div>
        </div>
      ) : events.length === 0 ? (
        <div className="no-posts-message">
          <h3>No events yet</h3>
          <p>Be the first to plan something awesome for your group!</p>
          <button
            onClick={() => setShowCreateEvent(true)}
            className="btn btn-primary"
          >
            Create First Event
          </button>
        </div>
      ) : (
        <div className="events-grid-modern">
          {events.map(event => (
            <div key={event.id} className="event-card-modern">
              {/* Event Card Header */}
              <div className="event-card-header">
                <div className="event-card-icon-wrapper">
                  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="3" y="4" width="18" height="18" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
                    <path d="M16 2v4M8 2v4M3 10h18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </div>
                <div className="event-card-date">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                    <path d="M12 6v6l4 2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                  {/* Prefer scheduled event_date when available, fallback to created_at */}
                  {(() => {
                    const d = event.event_date || event.eventDate || event.created_at || event.createdAt;
                    try {
                      return <span>{new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>;
                    } catch (e) {
                      return <span>—</span>;
                    }
                  })()}
                </div>
              </div>

              {/* Event Content */}
              <div className="event-card-body">
                <h3 className="event-card-title">{event.title}</h3>
                <p className="event-card-description">{event.description}</p>
              </div>

              {/* Voting Section */}
              <div className="event-voting-section">
                <h4 className="event-voting-title">Will you attend?</h4>
                
                <div className="event-voting-buttons">
                  <button
                    onClick={() => handleVote(event.id, 'yes')}
                    className={`event-vote-btn event-vote-yes ${event.user_vote === 'yes' ? 'active' : ''}`}
                  >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M20 6L9 17l-5-5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span className="vote-label">Yes</span>
                    <span className="vote-count">{event.yes_count || 0}</span>
                  </button>
                  
                  <button
                    onClick={() => handleVote(event.id, 'no')}
                    className={`event-vote-btn event-vote-no ${event.user_vote === 'no' ? 'active' : ''}`}
                  >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span className="vote-label">No</span>
                    <span className="vote-count">{event.no_count || 0}</span>
                  </button>
                </div>
                
                {event.user_vote && (
                  <div className="event-user-vote">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M22 11.08V12a10 10 0 11-5.93-9.14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      <path d="M22 4L12 14.01l-3-3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span>You voted: <strong>{event.user_vote === 'yes' ? 'Yes' : 'No'}</strong></span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}