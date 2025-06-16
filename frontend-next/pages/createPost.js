import Head from 'next/head';

export default function CreatePost() {
  return (
    <>
      <Head>
        <title>Create Post</title>
        <link rel="stylesheet" href="/css/style.css" />
        <script src="/js/createPost.js" defer></script>
        <script src="/js/session.js" defer></script>
        <script src="/js/socket.js" defer></script>
      </Head>
      <section id="createPostSection">
        <button className="return-button" onClick={() => window.location.href='/posts'}>Return</button>
        <div className="container-create">
          <form id="createPostForm">
            <h2>Create a New Post</h2>
            <p id="feedbackMessage" className="feedback-message"></p>
            <label htmlFor="title">Title:</label>
            <input type="text" id="title" required />
            <label htmlFor="content">Content:</label>
            <textarea id="content" required></textarea>
            <label htmlFor="category">Category:</label>
            <select id="category">
              <option value="General">General</option>
              <option value="Food">Food</option>
              <option value="Nature">Nature</option>
              <option value="Sport">Sport</option>
              <option value="Travel">Travel</option>
            </select>
            <label htmlFor="privacy">Privacy:</label>
            <select id="privacy">
              <option value="public">Public</option>
              <option value="almost_private">Almost Private</option>
              <option value="private">Private</option>
            </select>
            <button type="submit" className="button-main">Submit Post</button>
          </form>
        </div>
      </section>
    </>
  );
}
