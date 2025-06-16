import Head from 'next/head';
import { useEffect } from 'react';

export default function CreatePost() {
  useEffect(() => {
    const scripts = [
      "/js/createPost.js",
      "/js/socket.js",
      "/js/session.js"
    ];
    scripts.forEach(src => {
      const s = document.createElement("script");
      s.src = src;
      s.defer = true;
      document.body.appendChild(s);
    });
  }, []);

  return (
    <>
      <Head>
        <title>Welcome Page</title>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
      </Head>

      <section id="createPostSection">
        <button className="return-button" onClick={() => window.history.back()}>Return</button>
        <div className="container-create">
          <h1>Create a New Post</h1>

          <form id="createPostForm" method="POST">
            <label htmlFor="title">Title:</label>
            <input type="text" id="title" name="title" required />

            <label htmlFor="content">Content:</label>
            <textarea id="content" name="content" rows="6" required></textarea>

            <br /><br />
            <label>Category:</label><br />

            <label>
              <input type="checkbox" name="category" value="Nature" /> Nature
            </label>
            <label>
              <input type="checkbox" name="category" value="Food" /> Food
            </label>
            <label>
              <input type="checkbox" name="category" value="Sport" /> Sport
            </label>
            <label>
              <input type="checkbox" name="category" value="Travel" /> Travel
            </label>

            <br /><br />
            <input type="submit" value="Post it" className="button-create" />
          </form>
        </div>
      </section>
    </>
  );
}
