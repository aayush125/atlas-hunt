import type { Event } from '$lib/constants/wsmessages';

export type WSMessage = {
	type: Event;
	data: unknown;
};
