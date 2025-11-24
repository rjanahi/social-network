// pages/api/posts.js
export default async function handler(req, res) {
  try {
  const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
  let url = `${BACKEND_URL}/get-posts`;
    
    // Handle query parameters for category filtering
    if (req.query.category) {
      url += `?category=${encodeURIComponent(req.query.category)}`;
    }

    const response = await fetch(url, {
      method: req.method,
      headers: {
        'Content-Type': 'application/json',
        'Cookie': req.headers.cookie || '',
      },
    });
    
    // Get response text first to check if it's valid JSON
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
      // Backend returns posts as direct array, wrap it in posts property
      if (Array.isArray(data)) {
        data = { posts: data };
      }
    } catch (jsonError) {
      console.error('Response is not valid JSON:', responseText);
      // If not JSON, return an empty posts array or error message
      data = { posts: [], error: 'Invalid response from server' };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Posts API error:', error);
    res.status(500).json({ error: 'Failed to fetch posts', posts: [] });
  }
}