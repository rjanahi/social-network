import React, { useEffect } from 'react';

export default function Toast({ visible, message, type = 'info', onClose, duration = 3500 }) {
  useEffect(() => {
    if (!visible) return;
    const t = setTimeout(() => onClose && onClose(), duration);
    return () => clearTimeout(t);
  }, [visible, duration, onClose]);

  if (!visible) return null;

  const background = type === 'success' ? '#2b8a3e' : type === 'error' ? '#c53030' : '#2b6cb0';

  return (
    <div style={{
      position: 'fixed',
      bottom: '24px',
      left: '50%',
      transform: 'translateX(-50%)',
      zIndex: 9999,
      minWidth: '280px',
      maxWidth: '90%',
      boxShadow: '0 6px 18px rgba(0,0,0,0.12)'
    }}>
      <div style={{
        background,
        color: 'white',
        padding: '12px 16px',
        borderRadius: '8px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        fontSize: '14px'
      }}>
        <div style={{ marginRight: '12px', flex: 1 }}>{message}</div>
        <button onClick={() => onClose && onClose()} style={{
          background: 'transparent',
          border: 'none',
          color: 'rgba(255,255,255,0.9)',
          cursor: 'pointer',
          fontSize: '14px'
        }}>Close</button>
      </div>
    </div>
  );
}
