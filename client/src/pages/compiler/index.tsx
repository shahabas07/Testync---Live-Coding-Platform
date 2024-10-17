'use client'

import { useEffect, useState } from 'react';



const WebSocketComponent: React.FC = () => {
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [code, setCode] = useState<string>('');
    const [messages, setMessages] = useState<{ content: string }[]>([]);
    const [connectionStatus, setConnectionStatus] = useState<string>('Disconnected');
    const [mediaStream, setMediaStream] = useState<MediaStream | null>(null);

    useEffect(() => {
        const newSocket = new WebSocket('ws://localhost:8080/ws');

        newSocket.onopen = () => {
            setConnectionStatus('Connected');
            console.log('WebSocket connection established.');
        };

        newSocket.onmessage = (event) => {
            if (typeof event.data === 'string') {
                const message = JSON.parse(event.data);
                setMessages((prevMessages) => [...prevMessages, message]);
            } else {
                // Handle binary data (audio/video)
                console.log('Binary message received');
                const videoElement = document.getElementById('receivedVideo') as HTMLVideoElement;
                const blob = new Blob([event.data], { type: 'video/webm' });
                const url = URL.createObjectURL(blob);
                videoElement.src = url;
                videoElement.play();
            }
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

    const handleSendMessage = () => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ content: 'Message from user' }));
        }
    };

    // Capture and send audio/video stream
    const startMediaStream = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            setMediaStream(stream);

            const videoElement = document.getElementById('localVideo') as HTMLVideoElement;
            videoElement.srcObject = stream;
            videoElement.play();

            // Send video stream as binary data via WebSocket
            const mediaRecorder = new MediaRecorder(stream, { mimeType: 'video/webm' });
            mediaRecorder.ondataavailable = (event) => {
                if (socket && socket.readyState === WebSocket.OPEN) {
                    socket.send(event.data); // Send binary data
                }
            };
            mediaRecorder.start(1000); // Send data in chunks every second
        } catch (err) {
            console.error('Error capturing media stream:', err);
        }
    };

    return (
        // <div>hiihi</div>
        <div className="max-w-3xl mx-auto p-6 bg-gray-100 shadow-lg rounded-lg">
            <h1 className="text-3xl font-bold mb-4 text-center text-blue-700">WebSocket Code & Media Sharing</h1>
            <p className={`mb-4 text-center font-semibold ${connectionStatus === 'Connected' ? 'text-green-600' : 'text-red-600'}`}>
                Status: {connectionStatus}
            </p>
            
            {/* Code Editor */}
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

            {/* Media Stream Controls */}
            <div className="mt-6">
                <button
                    className="w-full px-4 py-2 bg-green-600 text-white font-semibold rounded-md hover:bg-green-700 transition duration-200 ease-in-out"
                    onClick={startMediaStream}
                >
                    Start Audio/Video Stream
                </button>
            </div>

            {/* Display Media */}
            <div className="mt-6 grid grid-cols-2 gap-4">
                {/* Local video stream */}
                <div>
                    <h2 className="text-xl font-semibold mb-2">Local Video:</h2>
                    <video id="localVideo" className="border border-gray-300 w-full h-64 bg-black" autoPlay muted></video>
                </div>
                {/* Received video stream */}
                <div>
                    <h2 className="text-xl font-semibold mb-2">Received Video:</h2>
                    <video id="receivedVideo" className="border border-gray-300 w-full h-64 bg-black" controls></video>
                </div>
            </div>

            {/* Display Messages */}
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