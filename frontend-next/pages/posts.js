import Head from 'next/head';

export default function Posts() {
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
      <section className="postsSection">
        <div className="container-post"></div>
      </section>
    </>
  );
}
