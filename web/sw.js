// Finance Tracker Service Worker (Stub)
// Implementasi lengkap untuk caching dan offline sync queue akan dilakukan pada fase selanjutnya.

self.addEventListener('install', (event) => {
    console.log('[Service Worker] Installed (Stub)');
    self.skipWaiting();
});

self.addEventListener('activate', (event) => {
    console.log('[Service Worker] Activated (Stub)');
    event.waitUntil(clients.claim());
});

self.addEventListener('fetch', (event) => {
    // Saat ini, pass-through (membiarkan semua request berjalan normal)
    // Di fase berikutnya, di sini kita akan mengimplementasikan cache-first atau network-first strategy
});
