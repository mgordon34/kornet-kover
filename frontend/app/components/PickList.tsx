// app/components/PicksList.tsx
import { PropPick } from '../types';
import PropPickItem from './PropPickItem'; // Import the StrategyItem component

interface PickListProps {
    picks: PropPick[];
}

const PickList: React.FC<PickListProps> = ({ picks }) => {
    return (
        <div className="border-b">
            <h1 className='text-3xl font-semibold p-4'>Strategy Picks</h1>
            <ul>
                {picks.map((pick) => (
                    <PropPickItem key={pick.id} pick={pick} />
                ))}
            </ul>
        </div>
    );
};

export default PickList;
