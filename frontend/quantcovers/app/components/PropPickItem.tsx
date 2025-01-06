// app/components/PropPickItem.tsx
import { PropPick } from '../types';

interface PropPickItemProps {
  pick: PropPick;
}

const PropPickItem: React.FC<PropPickItemProps> = ({ pick }) => {
  return (
    <li>
      <p className="border-b border-border/40 text-1xl p-2">[{pick.num_games}] {pick.player_name} {pick.side} {pick.line} {pick.stat} - Prediction: {pick.points} Diff: {pick.points - pick.line}</p>
    </li>
  );
};

export default PropPickItem;
