// app/strategies/page.tsx
import { Strategy } from '../types';
import StrategyList from '../components/StrategyList';

// Fetch strategies directly on the server side using `fetch`
async function getStrategies(): Promise<Strategy[]> {
  const userId = 1;
  const res = await fetch(`${process.env.API_URL}/strategies?user_id=${userId}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${process.env.JWT_TOKEN}`, // Or use session/cookies for auth
    },
    cache: "no-store",
  });
  return res.json();
}

const Strategies = async () => {
  const strategies = await getStrategies();

  return (
    <div className="items-center justify-items-center p-8">
      {/* Use the StrategyList component to display the strategies */}
      <StrategyList strategies={strategies} />
    </div>
  );
};

export default Strategies;

