// pages/api/get-otherPosts.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  const username = req.query.username || '';
  if (!username) {
    return res.status(400).json({ error: 'username required' });
  }

  try {
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/get-otherPosts/${encodeURIComponent(username)}`, {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
      },
    });

    const responseText = await response.text();
    let data;
    try {
      data = JSON.parse(responseText);
    } catch (jsonError) {
      console.error('get-otherPosts response is not valid JSON:', responseText);
      data = { profile: null, posts: [] };
    }

    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }

    res.status(response.status).json(data);
  } catch (error) {
    console.error('get-otherPosts API error:', error);
    res.status(500).json({ error: 'Failed to fetch user profile', profile: null, posts: [] });
  }
}
