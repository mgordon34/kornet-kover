// app/strategies/page.tsx
import { PropPick } from '../types';
import PickList from '../components/PickList';

// Fetch picks directly on the server side using `fetch`
async function getPicks(): Promise<PropPick[]> {
  const userId = 1;
  const res = await fetch(`${process.env.API_URL}/prop-picks?user_id=${userId}&date=2025-01-05`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${process.env.JWT_TOKEN}`, // Or use session/cookies for auth
    },
  });
  return res.json();
}

const Portfolio = async () => {
  const picks = await getPicks();

  return (
    <div className="items-center justify-items-center p-8">
      {/* Use the StrategyList component to display the strategies */}
      <PickList picks={picks} />
    </div>
  );
};

export default Portfolio;

