export type Treasure = {
	id: string;
	name: string;
	icon: string;
	x: number;
	y: number;
	question: {
		text: string;
		options: string[];
	};
};

export type TreasureItem = Omit<Treasure, 'x' | 'y' | 'question'>;
