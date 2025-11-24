export default async function handler(req, res) {
  if (req.method === 'POST') {
    try {
      const { groupId, userIds } = req.body;
      
      if (!groupId || !userIds || !Array.isArray(userIds) || userIds.length === 0) {
        return res.status(400).json({ error: 'Group ID and user IDs are required' });
      }

      const BACKEND_URL = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
      // Forward the request to the Go backend
      const response = await fetch(`${BACKEND_URL}/send-bulk-invitations`, {
        method: 'POST',
        headers: {
          'Cookie': req.headers.cookie || '',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          group_id: parseInt(groupId),
          user_ids: userIds.map(id => parseInt(id))
        })
      });

      const responseText = await response.text();
      let data;
      
      try {
        data = JSON.parse(responseText);
      } catch (parseError) {
        if (response.ok) {
          data = { success: true, message: 'Invitations sent successfully' };
        } else {
          data = { error: responseText || 'Failed to send invitations' };
        }
      }

      res.status(response.status).json(data);
    } catch (error) {
      console.error('Send invitations API error:', error);
      res.status(500).json({ error: 'Failed to send invitations' });
    }
  } else {
    res.setHeader('Allow', ['POST']);
    res.status(405).end(`Method ${req.method} Not Allowed`);
  }
}