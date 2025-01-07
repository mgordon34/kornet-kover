// app/components/StrategyItem.tsx
import { Strategy } from '../types';

interface StrategyItemProps {
  strategy: Strategy;
}

const StrategyItem: React.FC<StrategyItemProps> = ({ strategy }) => {
  return (
    <li>
      <p className="border-b border-border/40 text-1xl p-2">{strategy.id} {strategy.name}</p> {/* Render the strategy name */}
      {/* Add more strategy details here if necessary */}
    </li>
  );
};

export default StrategyItem;
