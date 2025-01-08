import { PropPick } from '../app/types';

export function getPrediction(pick: PropPick): number {
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

export function calculateDiff(pick: PropPick): number {
    const prediction = getPrediction(pick)

    return Math.abs(pick.line - prediction)
}
