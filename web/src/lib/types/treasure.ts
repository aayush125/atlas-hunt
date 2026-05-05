export type Treasure = {
	id: string;
	name: string;
	icon: string;
	rarity: string;
	points: number;
	boostPoints: number;
	x: number;
	y: number;
	question: {
		id: string;
		q: string;
		options: {
			A: string;
			B: string;
			C: string;
			D: string;
		};
	};
};

export type TreasureItem = Omit<Treasure, 'x' | 'y' | 'question'>;
