export const rarityToMultiplier: Record<string, number> = {
	common: 1,
	rare: 0.85,
	epic: 0.6,
	legendary: 0.55
};

export function getBonusPointsForPoints(points: number, count: number, multiplier: number): number {
	const factors = [1, 1.5, 2.5, 3.2, 4];
	const p = points;
	const m = multiplier;

	const round = (x: number): number => Math.ceil(Math.round(x * 1e2) / 1e2);

	if (count <= factors.length) {
		if (count === 1) return points;
		return round(p * factors[count - 1] * m);
	}

	const overflow = count - factors.length;
	return round(p * factors[factors.length - 1] * m + overflow * (p * 0.8 * m));
}

const RARITY_STYLES: Record<string, string> = {
	common: 'badge badge-sm badge-ghost font-semibold',
	rare: 'badge badge-sm badge-info font-bold',
	epic: 'badge badge-sm badge-secondary font-bold',
	legendary: 'badge badge-sm badge-warning font-extrabold'
};

export function getPlayerClaimedTreasureToastMessage(claimer: string, rarity: string): string {
	const rarityClass = RARITY_STYLES[rarity.toLowerCase()] ?? 'font-bold';

	return `<span class="font-semibold text-secondary">${claimer}</span> has claimed a <span class="${rarityClass}">${rarity}</span> item!`;
}
