// pages/api/groups/invitations.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/get-received-invitations`, {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
    });
    
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
      // Backend returns invitations as direct array, wrap it in invitations property
      if (Array.isArray(data)) {
        data = { invitations: data };
      }
    } catch (jsonError) {
      console.error('Invitations response is not valid JSON:', responseText);
      data = { invitations: [] };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Invitations API error:', error);
    res.status(500).json({ error: 'Failed to fetch invitations', invitations: [] });
  }
}