// pages/api/profile.js
export default async function handler(req, res) {
  try {
    const response = await fetch('http://localhost:8080/profile/complete', {
      method: 'GET',
      headers: {
        'Cookie': req.headers.cookie || '',
        'Content-Type': 'application/json',
      },
    });
    
    // Get response text first to check if it's valid JSON
    const responseText = await response.text();
    let data;
    
    try {
      data = JSON.parse(responseText);
    } catch (jsonError) {
      console.error('Profile response is not valid JSON:', responseText);
      // If not JSON, return empty profile
      data = { 
        profile: null,
        posts: [],
        followers: [],
        following: []
      };
    }
    
    // Forward any set-cookie headers from the backend
    const setCookieHeader = response.headers.get('set-cookie');
    if (setCookieHeader) {
      res.setHeader('Set-Cookie', setCookieHeader);
    }
    
    res.status(response.status).json(data);
  } catch (error) {
    console.error('Profile API error:', error);
    res.status(500).json({ 
      error: 'Failed to fetch profile',
      profile: null,
      posts: [],
      followers: [],
      following: []
    });
  }
}