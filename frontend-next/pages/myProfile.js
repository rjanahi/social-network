import Head from 'next/head';

export default function MyProfile() {
  return (
    <>
      <Head>
        <title>My Profile</title>
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
        <script src="/js/profile.js" defer></script>
      </Head>
      <section id="myProfileSection">
        <div className="profile-container">
          <div className="profile-header">
            <button className="return-button" onClick={() => window.location.href='/posts.html'}>Return</button>
            <h2>Welcome, <span id="profileUsername"></span></h2>
            <button onClick={() => window.location.href='/editProfile.html'}>Edit Profile</button>
          </div>
          <div className="profile-posts">
            <div className="container-myProfilePost"></div>
          </div>
        </div>
      </section>
    </>
  );
}
