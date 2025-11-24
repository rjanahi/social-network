// pages/api/groups/respond-request.js
export default async function handler(req, res) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Be permissive with payload shapes/casing
    const rawId = req.body?.request_id || req.body?.RequestID || req.body?.requestId || req.body?.id;
    const rawAction = (req.body?.action || req.body?.Action || req.body?.status || '').toString().toLowerCase();
    const request_id = Number(rawId || 0);
    const action = rawAction;

    if (!request_id || !action || !['accept', 'decline'].includes(action)) {
      return res.status(400).json({ error: 'Invalid request_id or action' });
    }

    // Convert action to status format expected by backend
    const status = action === 'accept' ? 'accepted' : 'declined';

    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';

    const response = await fetch(`${BACKEND_URL}/respond-join-request`, {
      method: 'POST',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        request_id: request_id,
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
        data = { success: true, message: `Request ${action}ed successfully` };
      } else {
        data = { error: responseText || `Failed to ${action} request` };
      }
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Respond request API error:', error);
    res.status(500).json({ error: 'Failed to process request' });
  }
}