// pages/api/categories.js
export default async function handler(req, res) {
  if (req.method !== 'GET') {
    res.setHeader('Allow', ['GET']);
    return res.status(405).end(`Method ${req.method} Not Allowed`);
  }

  try {
    // Return predefined categories since backend doesn't have a categories endpoint
    const categories = [
      'Travel',
      'Sport', 
      'Food',
      'Nature',
      'Technology',
      'Music',
      'Art',
      'Photography'
    ];
    
    res.status(200).json({ categories });
  } catch (error) {
    console.error('Categories API error:', error);
    res.status(500).json({ error: 'Failed to fetch categories', categories: [] });
  }
}