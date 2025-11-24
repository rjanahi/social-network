// pages/api/create-post.js
export const config = {
  api: {
    bodyParser: false,
  },
}

export default async function handler(req, res) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Get the raw body as a buffer
    const chunks = [];
    for await (const chunk of req) {
      chunks.push(chunk);
    }
    const buffer = Buffer.concat(chunks);

    // Create headers object and copy from request
    const headers = {};
    if (req.headers.cookie) {
      headers['Cookie'] = req.headers.cookie;
    }
    if (req.headers['content-type']) {
      headers['Content-Type'] = req.headers['content-type'];
    }
    if (req.headers['content-length']) {
      headers['Content-Length'] = req.headers['content-length'];
    }

    // Determine backend URL and forward the request to it
    const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    // Forward the request to the backend
    const response = await fetch(`${BACKEND_URL}/create-post`, {
      method: 'POST',
      body: buffer,
      headers: headers,
    });
    
    // Try to parse as JSON, fall back to text if it fails
    let data;
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      // Backend returned plain text (likely an error)
      const text = await response.text();
      data = { error: text, success: false };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Create post API error:', error);
    res.status(500).json({ error: 'Failed to create post' });
  }
}