// pages/api/groups/pending-approvals.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/get-pending-join-requests`, {
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
      // Backend returns requests as direct array, wrap it in requests property
      if (Array.isArray(data)) {
        data = { requests: data };
      }
    } catch (jsonError) {
      console.error('Pending approvals response is not valid JSON:', responseText);
      data = { requests: [] };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Pending approvals API error:', error);
    res.status(500).json({ error: 'Failed to fetch pending approvals', requests: [] });
  }
}