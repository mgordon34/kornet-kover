// app/strategies/page.tsx
import { StrategyPicks } from '../types';
import PickList from '../components/PickList';
import { calculateDiff } from '../../lib/pick_utils';

// Fetch picks directly on the server side using `fetch`
async function getPicks(): Promise<StrategyPicks[]> {
  const userId = 1;
  const formattedDate = new Intl.DateTimeFormat('en-CA').format(new Date());
  const res = await fetch(`${process.env.API_URL}/prop-picks?user_id=${userId}&date=${formattedDate}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${process.env.JWT_TOKEN}`, // Or use session/cookies for auth
    },
    cache: "no-store",
  });
  return res.json();
}

const Picks = async () => {
  const strategies = await getPicks();

  return (
    <div className="flex flex-col items-center justify-items-center p-8">
      <div className="inline-flex flex-col">
        {strategies.map((strategy) => {
          // Sort the picks for each strategy based on `calculateDiff`
          const sortedPicks = strategy.picks.sort((a, b) => calculateDiff(b) - calculateDiff(a));
          
          return (
            <div key={strategy.strat_id} className="my-4">
              <h2 className="text-xl font-bold">{strategy.strat_name}</h2>
              {/* Render the PickList component for each strategy */}
              <PickList picks={sortedPicks} />
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default Picks;

