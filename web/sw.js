// Bump on shell changes. A new worker waits for old tabs to close before activation.
const CACHE = 'foodweight-shell-v3';
const ASSETS = ['/assets/libre-baskerville-latin.woff2', '/assets/app.css', '/assets/theme.js', '/assets/shell.js', '/assets/calculator.js', '/assets/food-search.js', '/assets/food-prep.png', '/assets/icon.svg', '/assets/icon-192.png', '/assets/icon-512.png', '/manifest.webmanifest'];
const PAGES = ['/', '/?lang=en', '/?lang=it'];
self.addEventListener('install', event => {
    event.waitUntil(caches.open(CACHE).then(cache => cache.addAll([...ASSETS, ...PAGES])));
});
self.addEventListener('activate', event => {
    event.waitUntil(caches.keys().then(keys => Promise.all(keys.filter(key => key.startsWith('foodweight-shell-') && key !== CACHE).map(key => caches.delete(key)))));
});
self.addEventListener('fetch', event => {
    const url = new URL(event.request.url);
    // Never cache calculations, API responses, or authenticated/admin pages.
    if (event.request.method !== 'GET' || url.origin !== self.location.origin) return;
    const key = url.pathname + url.search;
    if (!ASSETS.includes(key) && !PAGES.includes(key)) return;
    event.respondWith((async () => {
        const cache = await caches.open(CACHE);
        try {
            const response = await fetch(event.request);
            if (response.ok) {
                // Storage quota errors must not discard a successful network response.
                try { await cache.put(event.request, response.clone()); } catch {}
            }
            return response;
        } catch {
            const cached = await cache.match(event.request);
            return cached || new Response('Offline', {status: 503});
        }
    })());
});
