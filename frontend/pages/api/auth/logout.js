// pages/api/auth/logout.js
export default async function handler(req, res) {
  if (req.method === 'POST') {
    try {
      // Determine backend URL
      const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
      // Forward the logout request to the Go backend
      const response = await fetch(`${BACKEND_URL}/logout`, {
        method: 'POST',
        headers: {
          'Cookie': req.headers.cookie || '',
        },
      });
      
      const data = await response.json();
      
      // Forward any set-cookie headers from the backend (likely to clear cookies)
      const setCookieHeader = response.headers.get('set-cookie');
      if (setCookieHeader) {
        res.setHeader('Set-Cookie', setCookieHeader);
      }
      
      res.status(response.status).json(data);
    } catch (error) {
      console.error('Logout failed:', error);
      res.status(500).json({ error: 'Logout failed' });
    }
  } else {
    res.setHeader('Allow', ['POST']);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}