import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { sveltekit } from '@sveltejs/kit/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			base: '/',
			injectRegister: 'auto',
			strategies: 'generateSW',
			srcDir: 'src',
			registerType: 'autoUpdate',
			includeAssets: ['favicon_io/favicon.ico', 'robots.txt', 'favicon_io/apple-touch-icon.png'],
			manifest: {
				name: 'Job Hunter',
				short_name: 'Job Hunter',
				start_url: '/',
				scope: '/',
				display: 'standalone',
				background_color: '#ecf3f8',
				theme_color: '#445def',
				icons: [
					{ src: '/favicon_io/android-chrome-192x192.png', sizes: '192x192', type: 'image/png' },
					{ src: '/favicon_io/android-chrome-512x512.png', sizes: '512x512', type: 'image/png' }
				]
			},
			workbox: {
				navigateFallbackDenylist: [/^\/api/, /^\/_/],
				globPatterns: ['**/*.{js,css,html,svg,png,ico,woff2}'],
				runtimeCaching: [
					{
						urlPattern: ({ request }) => request.destination === 'document',
						handler: 'NetworkFirst',
						options: { cacheName: 'pages', networkTimeoutSeconds: 4 }
					},
					{
						urlPattern: ({ request }) =>
							['style', 'script', 'worker'].includes(request.destination),
						handler: 'StaleWhileRevalidate',
						options: { cacheName: 'assets' }
					},
					{
						urlPattern: ({ request }) => request.destination === 'image',
						handler: 'CacheFirst',
						options: {
							cacheName: 'images',
							expiration: { maxEntries: 64, maxAgeSeconds: 60 * 60 * 24 * 30 }
						}
					}
				]
			},
			devOptions: {
				enabled: true
			}
		})
	],

	test: {
		expect: { requireAssertions: true },

		projects: [
			{
				extends: './vite.config.ts',

				test: {
					name: 'client',

					browser: {
						enabled: true,
						provider: playwright(),
						instances: [{ browser: 'chromium', headless: true }]
					},

					include: ['src/**/*.svelte.{test,spec}.{js,ts}']
					// exclude: ['src/lib/server/**']
				}
			},

			{
				extends: './vite.config.ts',

				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
