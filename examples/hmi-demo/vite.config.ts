import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

// The controller to talk to. Set CONTROLLER_URL to point the demo at any
// running nautilus controller (default: the local dashboard port).
//   CONTROLLER_URL=http://localhost:8080 npm run dev
export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	const controller = env.CONTROLLER_URL ?? 'http://localhost:8080';
	return {
		plugins: [sveltekit()],
		server: {
			// Proxy /api to the controller so the browser sees one origin.
			// changeOrigin rewrites Host but NOT the Origin header — the
			// controller's same-origin write guard would still see the vite
			// origin and 403 tag writes. The dev proxy genuinely fronts the
			// controller, so present the controller's own origin. (In
			// production this doesn't arise: the built SPA is served from
			// the controller itself, or writes carry a token.)
			proxy: {
				'/api': {
					target: controller,
					changeOrigin: true,
					configure(proxy) {
						proxy.on('proxyReq', (proxyReq) => {
							proxyReq.setHeader('origin', controller);
						});
					}
				}
			}
		}
	};
});
