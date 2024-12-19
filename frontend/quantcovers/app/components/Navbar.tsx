// app/components/Navbar.tsx
import Link from 'next/link';

const Navbar = () => {
  return (
    <nav className="bg-gray-800 text-white px-8 py-2 flex justify-between items-center">
      {/* Left Section */}
      <div className="flex space-x-8">
        <Link href="/" className="text-lg font-semibold hover:text-gray-400">
          Home
        </Link>
        <Link href="/strategies" className="text-lg font-semibold hover:text-gray-400">
          Strategies
        </Link>
      </div>

      {/* Right Section (Account) */}
      <div>
        <button className="text-lg font-semibold hover:text-gray-400">
          Account
        </button>
      </div>
    </nav>
  );
};


export default Navbar;

