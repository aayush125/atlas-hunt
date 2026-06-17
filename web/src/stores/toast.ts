import { writable } from 'svelte/store';

export type ToastType = 'info' | 'success' | 'warning' | 'error';

interface Toast {
	id: number;
	message: string;
	type: ToastType;
}

export const toasts = writable<Toast[]>([]);

let id = 0;

export function toast(message: string, type: ToastType = 'info', duration: number = 3000): void {
	const toastId = id++;
	toasts.update((t) => [...t, { id: toastId, message, type }]);
	setTimeout(() => {
		toasts.update((t) => t.filter((toast) => toast.id !== toastId));
	}, duration);
}
