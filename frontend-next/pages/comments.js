import Head from 'next/head';

export default function Comments() {
  return (
    <>
      <Head>
        <title>Comments</title>
        <link rel="stylesheet" href="/css/style.css" />
        <script src="/js/comments.js" defer></script>
        <script src="/js/likes.js" defer></script>
        <script src="/js/socket.js" defer></script>
        <script src="/js/session.js" defer></script>
      </Head>
      <section id="commentsSection"></section>
    </>
  );
}