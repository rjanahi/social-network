import Head from 'next/head';

export default function ErrorPage() {
  return (
    <>
      <Head>
        <title>Error</title>
        <link rel="stylesheet" href="/css/style.css" />
      </Head>
      <section id="errorSection">
        <div className="container-error">
          <h1>Oops! Something went wrong.</h1>
          <p>Please try again later or contact support.</p>
          <button className="return-button" onClick={() => window.location.href='/'}>Return Home</button>
        </div>
      </section>
    </>
  );
}
