// pages/api/group-details/[id].js
export default async function handler(req, res) {
  const { id } = req.query;

  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/group-details/${id}`, {
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
      console.error('Group details response is not valid JSON:', responseText);
      data = { error: 'Invalid response from server' };
    }
    
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Group details API error:', error);
    res.status(500).json({ error: 'Failed to fetch group details' });
  }
}