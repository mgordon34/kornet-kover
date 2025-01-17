import ClientPicks from "../components/ClientPicks"

export const dynamic = "force-dynamic"

const Picks = async () => {
  const fetchPicks = async (date: Date) => {
    const userId = 1;
    const formattedDate = new Intl.DateTimeFormat("en-CA").format(date);

    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/prop-picks?user_id=${userId}&date=${formattedDate}`, {
      method: "GET",
      cache: "no-store",
    });

    if (!res.ok) {
      throw new Error(`Failed to fetch picks: ${res.statusText}`);
    }

    return res.json();
  };

  // Initialize data for the current date
  const initialDate = new Date();
  const strategies = await fetchPicks(initialDate);

  return (
    <ClientPicks initialDate={initialDate} initialStrategies={strategies} />
  );
};

export default Picks;

