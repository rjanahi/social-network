// pages/api/handle-follow-request.js
export default async function handler(req, res) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Normalize incoming body to expected shape
    const rawRequester = req.body?.requester_id || req.body?.requesterId || req.body?.RequesterID || req.body?.From || req.body?.from;
    const rawAction = (req.body?.action || req.body?.Action || '').toString().toLowerCase();
    const payload = {
      requester_id: Number(rawRequester || 0),
      action: rawAction,
    };

    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const response = await fetch(`${BACKEND_URL}/handle-follow-request`, {
      method: 'POST',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    
    // Get response text first to check if it's valid JSON
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
    } catch (jsonError) {
      // If not JSON, treat the text as a message
      console.log('Handle follow request response (plain text):', responseText);
      data = { 
        message: responseText,
        success: response.ok
      };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Handle follow request API error:', error);
    res.status(500).json({ error: 'Failed to handle follow request' });
  }
}