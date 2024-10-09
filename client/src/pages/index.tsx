import Head from 'next/head';
import WebSocketComponent from '../components/WebSocketComponent';

const Home: React.FC = () => {
    return (
        <>
            <Head>
                <title>WebSocket Test</title>
                <meta name="description" content="WebSocket Test with Next.js" />
                <link rel="icon" href="/favicon.ico" />
            </Head>
            <main className="min-h-screen bg-gray-100 flex items-center justify-center">
                <WebSocketComponent />
            </main>
        </>
    );
};

export default Home;
