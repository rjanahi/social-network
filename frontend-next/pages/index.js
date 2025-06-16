import Head from 'next/head';

export default function MainPage() {
  return (
    <>
      <Head>
        <title>Main Page</title>
        <link rel="stylesheet" href="/css/style.css" />
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>
      <section id="mainSection">
        <div className="container-main">
          <div className="profile-top">
            <img src="/css/logo.png" alt="Logo" />
          </div>
          <div className="center-buttons">
            <button id="signUpButton" onClick={() => window.location.href='/signUp.html'}>Sign Up</button>
            <button id="logInButton" onClick={() => window.location.href='/login.html'}>Log In</button>
            <button id="postsButton" onClick={() => window.location.href='/posts.html'}>Enter Site</button>
            <button id="logoutButton" onClick={() => window.location.href='/logout.html'}>Log Out</button>
          </div>
        </div>
      </section>
    </>
  );
}