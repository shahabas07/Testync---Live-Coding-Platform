import { useEffect, useState } from 'react';

const WebSocketComponent: React.FC = () => {
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [code, setCode] = useState<string>('');
    const [messages, setMessages] = useState<{ content: string }[]>([]);
    const [connectionStatus, setConnectionStatus] = useState<string>('Disconnected');

    useEffect(() => {
        const newSocket = new WebSocket('ws://localhost:8080/ws');

        newSocket.onopen = () => {
            setConnectionStatus('Connected');
            console.log('WebSocket connection established.');
        };

        newSocket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            setMessages((prevMessages) => [...prevMessages, message]);
        };

        newSocket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        newSocket.onclose = () => {
            setConnectionStatus('Disconnected');
            console.log('WebSocket connection closed.');
        };

        setSocket(newSocket);

        return () => {
            newSocket.close();
        };
    }, []);

    const handleCodeChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
        const newCode = event.target.value;
        setCode(newCode);
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ content: newCode }));
        }
    };

    const handleSendMessage = () => { // Removed 'event' parameter
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ content: 'Message from user' }));
        }
    };    

    return (
        <div className="max-w-3xl mx-auto p-6 bg-gray-100 shadow-lg rounded-lg">
            <h1 className="text-3xl font-bold mb-4 text-center text-blue-700">WebSocket Code Editor</h1>
            <p className={`mb-4 text-center font-semibold ${connectionStatus === 'Connected' ? 'text-green-600' : 'text-red-600'}`}>
                Status: {connectionStatus}
            </p>
            <textarea
                className="w-full p-4 border border-gray-300 rounded-md mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500 transition duration-200 ease-in-out"
                rows={10}
                value={code}
                onChange={handleCodeChange}
                placeholder="Type your code here..."
            />
            <button
                className="w-full px-4 py-2 bg-blue-600 text-white font-semibold rounded-md hover:bg-blue-700 transition duration-200 ease-in-out"
                onClick={handleSendMessage}
            >
                Send Message
            </button>
            <div className="mt-6">
                <h2 className="text-xl font-semibold mb-2">Messages:</h2>
                <div className="border border-gray-300 p-2 rounded-md max-h-60 overflow-y-auto bg-white shadow-inner">
                    {messages.length === 0 ? (
                        <p className="text-gray-500 italic text-center">No messages yet.</p>
                    ) : (
                        messages.map((message, index) => (
                            <p key={index} className="text-gray-700 p-1 border-b last:border-b-0">{message.content}</p>
                        ))
                    )}
                </div>
            </div>
        </div>
    );
};

export default WebSocketComponent;
