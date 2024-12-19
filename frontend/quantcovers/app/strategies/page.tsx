// app/strategies/page.tsx
import { Strategy } from '../types';
import StrategyList from '../components/StrategyList';

// Fetch strategies directly on the server side using `fetch`
async function getStrategies(): Promise<Strategy[]> {
  const res = await fetch(`${process.env.NEXT_PUBLIC_BASE_URL}/api/strategies`);
  if (!res.ok) {
    throw new Error('Failed to fetch strategies');
  }
  return res.json();
}

const Strategies = async () => {
  const strategies = await getStrategies();

  return (
    <div>
      {/* Use the StrategyList component to display the strategies */}
      <StrategyList strategies={strategies} />
    </div>
  );
};

export default Strategies;

