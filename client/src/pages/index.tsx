import Link from 'next/link';

const Home: React.FC = () => {
    return (
        <div className="container mx-auto p-8 h-screen flex flex-col justify-between">
            {/* Top Section (Logo and Buttons) */}
            <div className="flex justify-between items-center mb-4">
                <div className="flex items-center">
                    <img 
                        src="https://i.ibb.co/9TLN3MH/Testync-removebg-preview.png"  
                        alt="Logo" 
                        className="h-10 w-10 mr-2"
                    />
                    <h1 className="text-2xl font-bold">Testync</h1>
                </div>
                <div className="flex items-center">
                    <button className="px-4 py-2 rounded-full bg-gray-200 text-gray-700 mr-2 border border-gray-800">
                        ?
                    </button>
                    <button className="px-4 py-2 rounded-md bg-gray-200 text-gray-700 mr-2">
                        lang
                    </button>
                    <button className="px-4 py-2 rounded-md bg-gray-200 text-gray-700">
                        BL
                    </button>
                </div>
            </div>

            

            <div className="flex-grow flex flex-col justify-center items-center">
            <div className="text-center mb-8">
                <h2 className="text-3xl font-bold mb-2">
                    Start your Live Coding Test
                </h2>
                <p className="text-lg">
                    Connect, collaborate from anywhere with Testync
                </p>
            </div>
                {/* Input and New Live Button */}
                <div className="flex justify-center mb-4">
                    <button className="px-8 py-3 rounded-full bg-gray-100 text-gray-700 mr-4">
                        + new Live
                    </button>
                    <input
                        type="text"
                        placeholder="enter live code..."
                        className="px-4 py-2 rounded-full bg-gray-100 text-gray-700"
                    />
                </div>

                {/* Join Button */}
                <Link href="/compiler">
                    <button className="px-8 py-3 rounded-full bg-blue-500 text-white text-center">
                        Join
                    </button>
                </Link>
            </div>
        </div>
    );
};

export default Home;
