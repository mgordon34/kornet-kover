// app/components/PropPickItem.tsx
import { PropPick } from '../types';

interface PropPickItemProps {
  pick: PropPick;
}

const PropPickItem: React.FC<PropPickItemProps> = ({ pick }) => {
  return (
    <li>
      <p className="border-b border-border/40 text-1xl p-2">{pick.id} {pick.playerIndex} {pick.side} {pick.line} {pick.stat}</p>
    </li>
  );
};

export default PropPickItem;
