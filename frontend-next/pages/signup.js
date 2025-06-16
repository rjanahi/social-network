import Head from 'next/head';

export default function SignUp() {
  return (
    <>
      <Head>
        <title>Sign Up</title>
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/signup.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>
      <section id="signUpSection">
        <div className="container-main">
          <h2>Registration Form</h2>
          <p style={{ color: 'red' }} id="feedbackMessage"></p>
          <form id="registrationForm" action="/signup" method="post">
            <p className="feedback-message" id="feedbackMessage"></p>
            <div className="form-group">
              <label htmlFor="username">Username:</label>
              <input type="text" id="username" name="username" required />
              <label htmlFor="fname">First name:</label>
              <input type="text" id="fname" name="fname" required />
              <label htmlFor="lname">Last name:</label>
              <input type="text" id="lname" name="lname" required />
              <label htmlFor="age">Age:</label>
              <input type="number" id="age" name="age" min="1" max="120" required />
            </div>
            <div className="form-group">
              <label htmlFor="gender">Gender:</label>
              <select id="gender" name="gender" required>
                <option value="" disabled>Select your gender</option>
                <option value="male">Male</option>
                <option value="female">Female</option>
              </select>
              <label htmlFor="email">Email:</label>
              <input type="email" id="email" name="email" placeholder="example@gmail.com" required />
              <label htmlFor="password">Password:</label>
              <input type="password" id="password" name="password" placeholder="Password" required />
            </div>
            <input type="submit" value="Submit" />
          </form>
        </div>
        <button className="return-button" onClick={() => window.location.href='/'}>Return</button>
      </section>
    </>
  );
}