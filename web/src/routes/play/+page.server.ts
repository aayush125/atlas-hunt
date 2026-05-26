import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const session = cookies.get('session_token');
	if (!session) {
		return { reconnectInfo: null };
	}

	const res = await fetch('/api/token/validate', {
		headers: {
			Cookie: `session_token=${session}`
		}
	});

	if (!res.ok) {
		return { reconnectInfo: null };
	}

	const json = await res.json();
	console.log('RESPONSE FROM GO SERVER', json);

	const reconnectInfo = json;
	return { reconnectInfo };
};
