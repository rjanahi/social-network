// pages/api/comments/[postId].js
export default async function handler(req, res) {
  const { postId } = req.query;

  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const response = await fetch(`http://localhost:8080/comments?post_id=${postId}`, {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
      },
    });
    
    const data = await response.json();
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Comments API error:', error);
    res.status(500).json({ error: 'Failed to fetch comments' });
  }
}