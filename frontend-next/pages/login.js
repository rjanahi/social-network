import Head from 'next/head'
import { useEffect } from 'react'

export default function Login() {
  return (
    <>
      <Head>
        <title>Welcome Page</title>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/login.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>

      <section id="logInSection">
        <div className="container-main">
          <h2>Log In Form</h2>
          <p id="logerror"></p>
          <form id="loginForm">
            <label htmlFor="userOremail">Enter your Username or Email:</label>
            <input type="text" id="userOremail" name="userOremail" required />

            <label htmlFor="pass">Password:</label>
            <input type="password" placeholder="Password" id="pass" name="pass" required />

            <br />
            <input type="submit" value="Submit" />
          </form>

          <p>Don't have an account?</p>
          <button id="signUpButtonLogin" className="button-main">Sign Up</button>
        </div>

        <button className="return-button" onClick={() => window.location.href = '/'}>Return</button>
      </section>
    </>
  )
}
