import Head from 'next/head';

export default function MainPage() {
  return (
    <>
      <Head>
        <title>Main Page</title>
        <link rel="stylesheet" href="/css/style.css" />
        <script src="/js/socket.js" defer></script>
        <script src="/js/main.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>
      <section id="mainSection">
        <div className="container-main">
          <div className="profile-top">
            <img src="/css/logo.png" alt="Logo" />
          </div>
          <div>
            <button id="signUpButton" className="button-main" onClick={() => window.location.href='/signUp'}>Sign Up</button>
            <button id="logInButton" className="button-main" onClick={() => window.location.href='/login'}>Log In</button>
            <button id="postsButton" className="button-main" onClick={() => window.location.href='/posts'}>Posts</button>
            <button id="logoutButton" className="button-main" onClick={() => window.location.href='/logout'}>Log Out</button>
          </div>
        </div>
      </section>
    </>
  );
}