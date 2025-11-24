// pages/api/user-followers.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const response = await fetch('http://localhost:8080/get-followers', {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      return res.status(response.status).json({ error: 'Failed to fetch followers' });
    }

    const data = await response.json();
    res.status(200).json(data);
  } catch (error) {
    console.error('User followers API error:', error);
    res.status(500).json({ error: 'Failed to fetch followers', followers: [] });
  }
}
