import Head from 'next/head';

export default function Login() {
  return (
    <>
      <Head>
        <title>Login</title>
        <link rel="stylesheet" href="/css/style.css" />
        <script src="/js/login.js" defer></script>
        <script src="/js/socket.js" defer></script>
      </Head>
      <section id="logInSection">
        <div className="container-login">
          <form id="loginForm">
            <h2>Login</h2>
            <p id="feedbackMessage" className="feedback-message"></p>
            <label>Email:</label>
            <input type="email" id="email" required />
            <label>Password:</label>
            <input type="password" id="password" required />
            <button type="submit" className="button-main">Log In</button>
            <p>Don't have an account? <a href="/signup">Sign up here</a></p>
          </form>
        </div>
      </section>
    </>
  );
}

