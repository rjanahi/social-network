// pages/api/session.js
export default async function handler(req, res) {
  if (req.method === 'GET') {
    try {
      // Determine backend URL from env (works in Docker Compose and local dev)
      const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
      // Forward the request to the Go backend
      const response = await fetch(`${BACKEND_URL}/check-session`, {
        method: 'GET',
        headers: {
          'Cookie': req.headers.cookie || '',
          'Content-Type': 'application/json',
        },
      });
      
      // Get response text first to check if it's valid JSON
      const responseText = await response.text();
      let data;
      
      try {
        data = JSON.parse(responseText);
      } catch (jsonError) {
        console.error('Session response is not valid JSON:', responseText);
        // If not JSON, assume not logged in
        data = { 
          loggedIn: false,
          userID: null,
          username: null,
          error: 'Invalid response from server'
        };
      }
      
      // Forward any cookies from the backend
      const setCookieHeader = response.headers.get('set-cookie');
      if (setCookieHeader) {
        res.setHeader('Set-Cookie', setCookieHeader);
      }
      
      // Forward the response
      res.status(response.status).json(data);
    } catch (error) {
      console.error('Session check failed:', error);
      res.status(500).json({ 
        error: 'Failed to check session',
        loggedIn: false,
        userID: null,
        username: null
      });
    }
  } else {
    res.setHeader('Allow', ['GET']);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}