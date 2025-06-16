import Head from 'next/head';

export default function Posts() {
  const handleCategoryClick = (value) => {
    const current = new URLSearchParams(window.location.search);
    current.set('category', value);
    window.location.search = current.toString();
  };

  const goTo = (path) => {
    window.location.href = path;
  };

  return (
    <>
      <Head>
        <title>Posts</title>
        <link rel="stylesheet" href="/css/style.css" />
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
        <script src="/js/posts.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/likes.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>

      <section id="postPageSection">
        <div className="sidebar-post left-sidebar">
          <br />
          <h2>Filter</h2>
          <br />
          <button className="button-side" onClick={() => document.getElementById('categoryOptions').classList.toggle('show')}>
            Categories
          </button>
          <div className="dropdown-post" id="categoryOptions">
            {['Travel', 'Sport', 'Food', 'Nature', 'Liked'].map((category) => (
              <button key={category} className="button-side" onClick={() => handleCategoryClick(category)}>
                {category}
              </button>
            ))}
          </div>
          <br />
          <button className="button-side" onClick={() => goTo('/mainPage')}>Main</button><br />
          <button id="postsButton" className="button-side">Posts</button><br />
          <button className="button-side" onClick={() => goTo('/myProfile')}>Profile</button><br />
          <button className="button-side" onClick={() => goTo('/chat')}>Chat</button><br />
          <button id="logoutButton" className="button-side" onClick={() => window.logout?.()}>Logout</button>
        </div>

        <div className="container-main">
          <button className="button-create" onClick={() => goTo('/createPost')}>Create Post</button>
        </div>

        <div className="container-post">
          <div id="postsContainer"></div>
        </div>
      </section>
    </>
  );
}
