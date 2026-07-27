import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	// Svelte's component inspector (dev only): hold Ctrl+Shift and hover to
	// see which component renders what; click to open it in your editor.
	// There's also a toggle button bottom-right of the page.
	vitePlugin: {
		inspector: {
			toggleKeyCombo: 'control-shift',
			holdMode: true,
			showToggleButton: 'always',
			toggleButtonPos: 'bottom-right'
		}
	},
	kit: {
		// A pure client SPA: it talks to a nautilus controller's tag API over
		// the network. `fallback` makes every route serve index.html, so the
		// built bundle drops into any static host (or beside the controller).
		adapter: adapter({ fallback: 'index.html' })
	}
};
