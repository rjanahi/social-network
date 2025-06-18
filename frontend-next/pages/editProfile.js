import Head from 'next/head';
import { useEffect } from 'react';

export default function EditProfile() {
 

  return (
    <>
      <Head>
        <title>Welcome Page</title>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/edit.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>

      <section id="editPageSection">
        <button className="return-button" onClick={() => window.location.href = '/myProfile'}>Return</button>

        <div className="container-main">
          <div className="profile-top">
            <img src="/css/logo.png" alt="Logo" />
          </div>
          <h2>Edit Profile</h2>
          <p id="editProfileMessage" className="message"></p>
          <form id="editProfileForm">
            <p className="feedback-message" id="feedbackMessage"></p>

            <div className="form-group">
              <label htmlFor="bio">Bio:</label>
              <textarea id="bio" name="bio"></textarea>

              <label htmlFor="username">Username:</label>
              <input type="text" id="username" name="username" />

              <label htmlFor="fname">First name:</label>
              <input type="text" id="fname" name="fname" required />

              <label htmlFor="lname">Last name:</label>
              <input type="text" id="lname" name="lname" />

              <label htmlFor="age">Age:</label>
              <input type="number" id="age" name="age" min="1" max="120" />
            </div>

            <div className="form-group">
              <label htmlFor="gender">Gender:</label>
              <select id="gender" name="gender">
                <option value="" disabled>Select your gender</option>
                <option value="male">Male</option>
                <option value="female">Female</option>
              </select>

              <label htmlFor="email">Email:</label>
              <input type="email" id="email" name="email" placeholder="example@gmail.com" />

              <label htmlFor="password">Password:</label>
              <input type="password" placeholder="Password" id="password" name="password" />
            </div>

            <button type="submit" className="button-main">Update Profile</button>
          </form>
        </div>
      </section>
    </>
  );
}
