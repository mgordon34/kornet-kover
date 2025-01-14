// app/strategies/page.tsx
import { PropPick } from '../types';
import PickList from '../components/PickList';
import { calculateDiff } from '../../lib/pick_utils';

// Fetch picks directly on the server side using `fetch`
async function getPicks(): Promise<PropPick[]> {
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

const Portfolio = async () => {
  const picks = await getPicks();

  const p1 = picks.filter(x => x.strat_id == 1).sort((a,b) => calculateDiff(b) - calculateDiff(a));
  const p2 = picks.filter(x => x.strat_id == 2).sort((a,b) => calculateDiff(b) - calculateDiff(a));
  return (
    <div className="flex flex-col items-center justify-items-center p-8">
      <div className="inline-flex flex-col">
        <PickList picks={p1} />
        <PickList picks={p2} />
      </div>
    </div>
  );
};

export default Portfolio;

