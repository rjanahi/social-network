export default async function handler(req, res) {
  if (req.method === 'POST') {
    try {
      // For now, let's just proxy the request directly to backend
      const { group_id, title, content } = req.body;
      
      const formData = new FormData();
      formData.append('group_id', group_id);
      formData.append('title', title);
      formData.append('content', content);

      const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
      const response = await fetch(`${BACKEND_URL}/groups/create-post`, {
        method: 'POST',
        headers: {
          'Cookie': req.headers.cookie || ''
        },
        body: formData
      });

      if (response.ok) {
        const data = await response.json();
        res.status(200).json(data);
      } else {
        const errorText = await response.text();
        res.status(response.status).json({ error: errorText });
      }
    } catch (error) {
      console.error('API Error:', error);
      res.status(500).json({ error: 'Internal server error' });
    }
  } else {
    res.setHeader('Allow', ['POST']);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}