import Head from 'next/head';
import { useEffect } from 'react';

export default function MyProfile() {
 

  return (
    <>
      <Head>
        <title>Welcome Page</title>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
        <script src="/js/posts.js" defer></script>
        <script src="/js/likes.js" defer></script>
        <script src="/js/comments.js" defer></script>
      </Head>

      <section id="profilePageSection">
        <button className="return-button" onClick={() => window.location.href = '/posts'}>Return</button>

        <div className="container-main">
          <div className="profile-top">
            <img src="/css/logo.png" alt="Logo" />
            <div className="follow">
              <p id="myProfileUsername">Username</p>
              <p>Followers: <span id="userFollowers">0</span></p>
              <p>Following: <span id="userFollowing">0</span></p>
            </div>
          </div>

          <button id="editProfileButton" onClick={() => window.location.href = '/editProfile'} className="button-main">
            Edit Profile
          </button>

          <div>
            <p id="profileBio" className="bio"></p>
          </div>

          <section id="postPageSection">
            <button className="button-create" onClick={() => window.location.href = '/createPost'}>Create Post</button>
            <div className="container-profilePost">
              <div id="postsContainer"></div>
            </div>
          </section>
        </div>
      </section>
    </>
  );
}
