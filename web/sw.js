const CACHE_NAME = 'finance-tracker-v13'; // Ubah angka versi ini setiap kali ada pembaruan kode/UI

// Daftar semua aset statis yang perlu di-cache saat proses instalasi PWA
const PRECACHE_ASSETS = [
    '/',
    '/index.html',
    '/manifest.json',
    '/static/css/pico.min.css',
    '/static/js/alpine.min.js',
    '/static/js/app.js',
    '/static/js/chart.min.js',
    '/static/icons/icon-192x192.png',
    '/static/icons/icon-512x512.png',
    '/static/icons/apple-touch-icon.png',
    '/static/icons/base-icon.svg'
];

// ==========================================
// LIFECYCLE: INSTALL
// ==========================================
self.addEventListener('install', (event) => {
    console.log('[Service Worker] Installing Service Worker...', CACHE_NAME);
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => {
            console.log('[Service Worker] Pre-caching static assets');
            return cache.addAll(PRECACHE_ASSETS);
        })
    );
    // Memaksa service worker baru untuk langsung aktif tanpa menunggu halaman ditutup
    self.skipWaiting();
});

// ==========================================
// LIFECYCLE: ACTIVATE
// ==========================================
self.addEventListener('activate', (event) => {
    console.log('[Service Worker] Activating Service Worker...');
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    // Hapus cache lama jika namanya tidak cocok dengan CACHE_NAME saat ini
                    if (cacheName !== CACHE_NAME) {
                        console.log('[Service Worker] Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        })
    );
    // Segera ambil alih kontrol atas semua client yang terbuka
    event.waitUntil(self.clients.claim());
});

// ==========================================
// FETCH HANDLER (CACHING STRATEGIES)
// ==========================================
self.addEventListener('fetch', (event) => {
    // Hanya proses method GET. (POST, PUT, DELETE akan di-handle offline sync terpisah nantinya)
    if (event.request.method !== 'GET') return;

    const requestUrl = new URL(event.request.url);

    // ----------------------------------------------------------------------
    // STRATEGI A: API REQUEST (Network First + Cache Fallback)
    // Berlaku jika URL request mengandung path '/api/v1/'
    // ----------------------------------------------------------------------
    if (requestUrl.pathname.includes('/api/v1/')) {
        event.respondWith(
            fetch(event.request)
                .then((networkResponse) => {
                    // Jika sukses dari jaringan, simpan salinannya ke cache agar selalu ter-update
                    return caches.open(CACHE_NAME).then((cache) => {
                        cache.put(event.request, networkResponse.clone());
                        return networkResponse;
                    });
                })
                .catch(async () => {
                    // Jika jaringan terputus (offline), ambil data terakhir dari cache
                    console.log('[Service Worker] API Fetch failed, falling back to cache:', event.request.url);
                    const cachedResponse = await caches.match(event.request);
                    if (cachedResponse) {
                        return cachedResponse;
                    }
                    
                    // Jika cache juga kosong, kembalikan response JSON fallback
                    return new Response(JSON.stringify({ 
                        error: 'Anda sedang offline dan tidak ada data cache yang tersedia.',
                        fallback: true 
                    }), {
                        status: 503,
                        headers: { 'Content-Type': 'application/json' }
                    });
                })
        );
        return; // Hentikan eksekusi handler untuk API
    }

    // ----------------------------------------------------------------------
    // STRATEGI B: STATIC ASSETS (Cache First + Network Fallback)
    // Berlaku untuk gambar, CSS, JS, dan HTML
    // ----------------------------------------------------------------------
    event.respondWith(
        caches.match(event.request)
            .then((cachedResponse) => {
                // Jika aset ditemukan di cache, kembalikan seketika!
                if (cachedResponse) {
                    return cachedResponse;
                }

                // Jika aset tidak ada di cache, unduh via jaringan
                return fetch(event.request)
                    .then((networkResponse) => {
                        // Jangan caching respons yang tidak valid
                        if (!networkResponse || networkResponse.status !== 200 || networkResponse.type !== 'basic') {
                            return networkResponse;
                        }

                        // Simpan respons baru ke cache untuk pemakaian berikutnya
                        const responseToCache = networkResponse.clone();
                        caches.open(CACHE_NAME).then((cache) => {
                            cache.put(event.request, responseToCache);
                        });

                        return networkResponse;
                    })
                    .catch((err) => {
                        console.error('[Service Worker] Fetch for static asset failed:', event.request.url, err);
                        // Kita bisa menambahkan fallback custom offline.html di sini jika mau
                    });
            })
    );
});
