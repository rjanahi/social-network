import Head from 'next/head';

export default function ErrorPage() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <Head>
        <title>Error - SocialNet</title>
      </Head>
      <div className="max-w-md w-full text-center">
        <div className="bg-white p-8 rounded-lg shadow-lg border border-gray-200">
          <div className="text-red-500 text-6xl mb-4 font-bold">ERROR</div>
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Oops! Something went wrong</h1>
          <p className="text-gray-600 mb-6">Please try again later or contact support if the problem persists.</p>
          <button 
            className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors"
            onClick={() => window.location.href='/'}
          >
            Return Home
          </button>
        </div>
      </div>
    </div>
  );
}
