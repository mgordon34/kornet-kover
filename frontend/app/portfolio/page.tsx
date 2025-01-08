// app/strategies/page.tsx
import { PropPick } from '../types';
import PickList from '../components/PickList';

// Fetch picks directly on the server side using `fetch`
async function getPicks(): Promise<PropPick[]> {
  const userId = 1;
  const now = new Date();
  console.log(process.env.API_URL)
  const res = await fetch(`${process.env.API_URL}/prop-picks?user_id=${userId}&date=${now.toISOString().split('T')[0]}`, {
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

  const p1 = picks.filter(x => x.strat_id == 1);
  const p2 = picks.filter(x => x.strat_id == 2);
  return (
    <div className="items-center justify-items-center p-8">
      <PickList picks={p1} />
      <PickList picks={p2} />
    </div>
  );
};

export default Portfolio;

