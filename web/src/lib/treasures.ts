import raw from './assets/treasures.json';

export const placeHolderTreasures = [
	{ id: '1', name: 'golden-crown', icon: '👑' },
	{ id: '2', name: 'ruby-gem', icon: '💎' },
	{ id: '3', name: 'ancient-map', icon: '🗺️' },
	{ id: '4', name: 'gold-coin', icon: '🪙' },
	{ id: '5', name: 'magic-wand', icon: '🪄' },
	{ id: '6', name: 'crystal-ball', icon: '🔮' },
	{ id: '7', name: 'treasure-chest', icon: '📦' },
	{ id: '8', name: 'ancient-key', icon: '🗝️' },
	{ id: '9', name: 'enchanted-ring', icon: '💍' },
	{ id: '10', name: 'sacred-scroll', icon: '📜' },
	{ id: '11', name: 'golden-trophy', icon: '🏆' },
	{ id: '12', name: 'mystic-potion', icon: '🧪' },
	{ id: '13', name: 'lucky-clover', icon: '🍀' },
	{ id: '14', name: 'pirates-compass', icon: '🧭' },
	{ id: '15', name: 'dragon-egg', icon: '🥚' },
	{ id: '16', name: 'shining-star', icon: '⭐' }
];

export const rawTreasures = raw;
export const treasuresById = new Map(raw.map((item) => [item.id, item]));

export function rarityClass(rarity: string) {
	switch (rarity.toLowerCase()) {
		case 'common':
			return 'badge-ghost';
		case 'uncommon':
			return 'badge-info';
		case 'rare':
			return 'badge-secondary';
		case 'legendary':
			return 'badge-warning';
		default:
			return 'badge-outline';
	}
}
