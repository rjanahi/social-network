// pages/api/users.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const { search } = req.query;
    
    // Build query parameters
    const params = new URLSearchParams();
    if (search) {
      params.append('search', search);
    }
    
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/all-users?${params}`, {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
      },
    });
    
    // Get response text first to check if it's valid JSON
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
    } catch (jsonError) {
      console.error('Users response is not valid JSON:', responseText);
      // If not JSON, return empty users array
      data = { users: [] };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Users API error:', error);
    res.status(500).json({ error: 'Failed to fetch users' });
  }
}