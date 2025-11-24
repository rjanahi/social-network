// pages/api/group-messages.js
export default async function handler(req, res) {
  if (req.method === 'GET') {
    try {
      const { group_id, offset = 0 } = req.query;
      
      if (!group_id) {
        return res.status(400).json({ error: 'Missing group_id parameter' });
      }

      const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
      // Forward the request to the Go backend
      const response = await fetch(`${BACKEND_URL}/group-messages?group_id=${group_id}&offset=${offset}`, {
        method: 'GET',
        headers: {
          'Cookie': req.headers.cookie || '',
          'Content-Type': 'application/json',
        },
      });

      const data = await response.json();

      // Forward any cookies from the backend
      const setCookieHeader = response.headers.get('set-cookie');
      if (setCookieHeader) {
        res.setHeader('Set-Cookie', setCookieHeader);
      }

      // Forward the response
      res.status(response.status).json(data);
    } catch (error) {
      console.error('Group messages fetch failed:', error);
      res.status(500).json({ 
        error: 'Failed to fetch group messages'
      });
    }
  } else {
    res.setHeader('Allow', ['GET']);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}