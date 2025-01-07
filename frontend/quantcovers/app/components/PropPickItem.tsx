// app/components/PropPickItem.tsx
import { PropPick } from '../types';

interface PropPickItemProps {
  pick: PropPick;
}

function getPrediction(pick: PropPick): number {
    switch (pick.stat) {
        case "points":
            return pick.points;
        case "rebounds":
            return pick.rebounds;
        case "assists":
            return pick.assists;

        default:
            return pick.points;
    }
}

function calculateDiff(pick: PropPick): number {
    const prediction = getPrediction(pick)

    return Math.abs(pick.line - prediction)
}

const PropPickItem: React.FC<PropPickItemProps> = ({ pick }) => {
  return (
    <li>
      <p className="border-b border-border/40 text-1xl p-2">[{pick.num_games}] {pick.player_name} {pick.side} {pick.line} {pick.stat} - Prediction: {getPrediction(pick).toFixed(2)} Diff: {calculateDiff(pick).toFixed(2)}</p>
    </li>
  );
};

export default PropPickItem;
