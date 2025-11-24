// pages/api/groups/respond-invitation.js
export default async function handler(req, res) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Be permissive with incoming payloads (many clients may send different casing or types)
    const rawId = req.body?.invitation_id || req.body?.InvitationID || req.body?.invitationId || req.body?.id;
    const rawAction = (req.body?.action || req.body?.Action || req.body?.status || '').toString().toLowerCase();
    let invitation_id = Number(rawId || 0);
    const action = rawAction;

    if (!action || !['accept', 'decline'].includes(action)) {
      return res.status(400).json({ error: 'Invalid action' });
    }

    // If client didn't provide invitation_id, try to resolve it server-side by
    // fetching pending invitations for the current user and matching by
    // group_id + inviter (many websocket payloads include group and inviter but
    // not the DB id). This avoids client-side fragile matching and prevents
    // the UI from throwing 'Invitation id not found'.
    if (!invitation_id) {
      try {
        // Allow permissive casing for incoming match fields
        const gid = req.body?.group_id || req.body?.GroupID || req.body?.groupId || req.body?.group || null;
        const inviter = req.body?.inviter_id || req.body?.InviterID || req.body?.inviterId || req.body?.from || req.body?.From || null;

        const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
        const invResp = await fetch(`${BACKEND_URL}/get-received-invitations`, {
          method: 'GET',
          headers: { 'Cookie': req.headers.cookie || '' },
        });
        if (invResp.ok) {
          const invs = await invResp.json();
          // invs is expected to be an array of invitation objects with id, group_id, inviter_id
          const match = Array.isArray(invs) ? invs.find(i => {
            const okGroup = gid == null || String(i.group_id) == String(gid);
            const okInviter = inviter == null || String(i.inviter_id) == String(inviter);
            return okGroup && okInviter;
          }) : null;
          if (match) {
            invitation_id = Number(match.id || 0);
          }
        }
      } catch (e) {
        console.error('Failed to resolve invitation id server-side', e);
      }
    }

    if (!invitation_id) {
      return res.status(400).json({ error: 'invitation_id not provided and could not be resolved' });
    }

    // Convert action to status format expected by backend
    const status = action === 'accept' ? 'accepted' : 'declined';

    const response = await fetch(`${BACKEND_URL}/respond-group-invitation`, {
      method: 'POST',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        invitation_id: invitation_id,
        status: status
      }),
    });
    
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
    } catch (jsonError) {
      console.error('Response is not valid JSON:', responseText);
      if (response.ok) {
        data = { success: true, message: `Invitation ${action}ed successfully` };
      } else {
        data = { error: responseText || `Failed to ${action} invitation` };
      }
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Respond invitation API error:', error);
    res.status(500).json({ error: 'Failed to process invitation response' });
  }
}