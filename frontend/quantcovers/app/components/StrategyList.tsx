// app/components/StrategyList.tsx
import { Strategy } from '../types';
import StrategyItem from './StrategyItem'; // Import the StrategyItem component

interface StrategyListProps {
  strategies: Strategy[];
}

const StrategyList: React.FC<StrategyListProps> = ({ strategies }) => {
  return (
    <div>
      <h1 className='text-3xl'>Your Strategies</h1>
      <ul>
        {strategies.map((strategy) => (
          <StrategyItem key={strategy.id} strategy={strategy} />
        ))}
      </ul>
    </div>
  );
};

export default StrategyList;
