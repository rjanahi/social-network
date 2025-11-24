import "@/styles/index.css";
import { WebSocketProvider } from '../contexts/WebSocketContext';
import { useSession } from '../hooks/useSession';

function AppContent({ Component, pageProps }) {
  const { user } = useSession();
  
  return (
    <WebSocketProvider userID={user?.userID}>
      <Component {...pageProps} />
    </WebSocketProvider>
  );
}

export default function App(props) {
  return <AppContent {...props} />;
}
