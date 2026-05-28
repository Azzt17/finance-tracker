// Register Service Worker for PWA
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(reg => {
                console.log('ServiceWorker registered with scope:', reg.scope);
                
                // Deteksi jika ada versi Service Worker baru (Deploy Versi Baru)
                reg.onupdatefound = () => {
                    const installingWorker = reg.installing;
                    if (installingWorker == null) return;
                    
                    installingWorker.onstatechange = () => {
                        if (installingWorker.state === 'installed' && navigator.serviceWorker.controller) {
                            // Versi baru ditemukan dan sudah diinstal
                            console.log('New version available! Updating...');
                            
                            // Akses store Alpine untuk memunculkan Toast
                            const financeStore = document.querySelector('[x-data]')?._x_dataStack?.[0]?.$store?.finance;
                            if (financeStore) {
                                financeStore.showFeedback("✨ Versi baru diunduh. Memuat ulang...", false);
                            }
                            
                            // Karena sw.js kita menggunakan self.skipWaiting(),
                            // worker baru akan langsung aktif. Kita hanya perlu me-refresh halaman 
                            // agar client memakai cache/aset baru.
                            setTimeout(() => {
                                window.location.reload();
                            }, 2000);
                        }
                    };
                };
            })
            .catch(err => console.error('ServiceWorker registration failed:', err));
    });
}

document.addEventListener('alpine:init', () => {
    Alpine.store('finance', {
        // ==========================================
        // STATE
        // ==========================================
        currentView: 'input', // Valid values: 'input' | 'dashboard' | 'history' | 'settings'
        
        // Dynamic current month (YYYY-MM)
        currentMonth: new Date().toISOString().slice(0, 7),
        
        budget: { total_budget: 0, total_spent: 0, remaining_balance: 0 },
        savingsGoals: [],
        recentTransactions: [],
        categories: [],
        
        isLoading: false,
        isOffline: !navigator.onLine,
        pendingSync: 0,
        
        feedback: { show: false, message: '', isError: false },
        
        db: null, // IndexedDB instance

        // ==========================================
        // GETTERS / COMPUTED
        // ==========================================
        get quickAddCategories() {
            // Returns categories where is_quick_add is true (or 1), sorted by sort_order
            return this.categories
                .filter(c => c.is_quick_add === true || c.is_quick_add === 1)
                .sort((a, b) => a.sort_order - b.sort_order);
        },

        // ==========================================
        // ACTIONS / METHODS
        // ==========================================
        async init() {
            // Register event listeners for network status on window
            window.addEventListener('online', () => this.updateOfflineStatus(false));
            window.addEventListener('offline', () => this.updateOfflineStatus(true));

            // Initialize IndexedDB
            await this.initDB();

            // Load initial app data
            this.loadInitialData();
        },

        // ==========================================
        // INDEXEDDB LOGIC (OFFLINE STORAGE)
        // ==========================================
        async initDB() {
            return new Promise((resolve, reject) => {
                const request = indexedDB.open('finance_db', 1);
                
                request.onupgradeneeded = (event) => {
                    const db = event.target.result;
                    if (!db.objectStoreNames.contains('sync_queue')) {
                        db.createObjectStore('sync_queue', { keyPath: 'client_transaction_id' });
                    }
                };

                request.onsuccess = (event) => {
                    this.db = event.target.result;
                    this.checkPendingSync();
                    resolve(this.db);
                };

                request.onerror = (event) => {
                    console.error('IndexedDB error:', event.target.error);
                    reject(event.target.error);
                };
            });
        },

        async saveToSyncQueue(transaction) {
            return new Promise((resolve, reject) => {
                if (!this.db) return reject('DB not initialized');
                const tx = this.db.transaction('sync_queue', 'readwrite');
                const store = tx.objectStore('sync_queue');
                const req = store.put(transaction);
                req.onsuccess = () => resolve();
                req.onerror = () => reject(req.error);
            });
        },

        async getSyncQueue() {
            return new Promise((resolve, reject) => {
                if (!this.db) return resolve([]);
                const tx = this.db.transaction('sync_queue', 'readonly');
                const store = tx.objectStore('sync_queue');
                const req = store.getAll();
                req.onsuccess = () => resolve(req.result);
                req.onerror = () => reject(req.error);
            });
        },

        async clearSyncQueue() {
            return new Promise((resolve, reject) => {
                if (!this.db) return resolve();
                const tx = this.db.transaction('sync_queue', 'readwrite');
                const store = tx.objectStore('sync_queue');
                const req = store.clear();
                req.onsuccess = () => resolve();
                req.onerror = () => reject(req.error);
            });
        },

        async checkPendingSync() {
            const queue = await this.getSyncQueue();
            this.pendingSync = queue.length;
        },

        async syncTransactions() {
            if (this.isOffline) return;
            
            const queue = await this.getSyncQueue();
            if (queue.length === 0) return;

            try {
                this.showFeedback("Mulai sinkronisasi data...", false);
                const res = await fetch('/api/v1/sync/transactions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ transactions: queue })
                });

                if (res.ok) {
                    await this.clearSyncQueue();
                    this.pendingSync = 0;
                    await this.loadInitialData(); // Refresh UI without pending local data
                    this.showFeedback(`✓ Sinkronisasi berhasil (${queue.length} transaksi)`, false);
                } else {
                    console.error("Gagal sinkronisasi data");
                    this.showFeedback("❌ Gagal sinkronisasi", true);
                }
            } catch (err) {
                console.error("Sinkronisasi tertunda (jaringan bermasalah)", err);
            }
        },

        // ==========================================
        // UI AND DATA FETCHING
        // ==========================================
        changeView(viewName) {
            const validViews = ['input', 'dashboard', 'history', 'settings'];
            if (validViews.includes(viewName)) {
                this.currentView = viewName;
            } else {
                console.warn(`View tidak valid: ${viewName}`);
            }
        },

        getCategoryName(id) {
            if (!id) return "Uncategorized";
            const cat = this.categories.find(c => c.id === id);
            return cat ? `${cat.icon_emoji || ''} ${cat.name}`.trim() : "Unknown";
        },

        async loadInitialData() {
            this.isLoading = true;
            try {
                const month = this.currentMonth;
                
                // Fetch all data in parallel
                const [cats, budg, sav, txs] = await Promise.all([
                    fetch('/api/v1/categories').then(r => r.ok ? r.json() : []),
                    fetch(`/api/v1/budget/${month}`).then(r => r.ok ? r.json() : { total_budget: 0, total_spent: 0, remaining_balance: 0 }),
                    fetch(`/api/v1/savings?year_month=${month}`).then(r => r.ok ? r.json() : []),
                    fetch(`/api/v1/transactions?year_month=${month}`).then(r => r.ok ? r.json() : [])
                ]);

                const pendingTxs = await this.getSyncQueue();

                this.categories = cats || [];
                this.savingsGoals = sav || [];
                
                // Optimistic UI: Merge remote transactions with pending offline transactions
                this.recentTransactions = [...pendingTxs, ...(txs || [])].sort((a, b) => new Date(b.transacted_at) - new Date(a.transacted_at));
                
                // Optimistic UI: Adjust budget based on pending transactions
                let pendingSpent = pendingTxs.reduce((sum, tx) => sum + tx.amount, 0);
                let actualBudget = budg || { total_budget: 0, total_spent: 0, remaining_balance: 0 };
                
                this.budget = {
                    total_budget: actualBudget.total_budget,
                    total_spent: actualBudget.total_spent + pendingSpent,
                    remaining_balance: actualBudget.remaining_balance - pendingSpent
                };
                
                console.log('Initial data loaded successfully');
            } catch (error) {
                console.error('Failed to load initial data', error);
            } finally {
                this.isLoading = false;
            }
        },

        async submitTransaction(amount, note, categoryId) {
            if (!amount) return;
            
            const payload = {
                client_transaction_id: crypto.randomUUID(),
                amount: parseInt(amount, 10),
                category_id: categoryId || null,
                note: note || '',
                transacted_at: new Date().toISOString()
            };

            if (this.isOffline) {
                // Store in IndexedDB and sync later
                await this.saveToSyncQueue(payload);
                await this.checkPendingSync();
                
                // Refresh UI optimistically
                await this.loadInitialData();
                this.showFeedback("✓ Tersimpan offline di antrean", false);
                return;
            }

            try {
                const res = await fetch('/api/v1/transactions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                
                if (res.ok) {
                    await this.loadInitialData(); // Refresh data
                    this.showFeedback("✓ Tercatat", false);
                } else {
                    const err = await res.json();
                    this.showFeedback("Gagal: " + err.error, true);
                }
            } catch (err) {
                console.error('Submit transaction error:', err);
                
                // Fallback to offline queue if network fails completely despite being "online"
                await this.saveToSyncQueue(payload);
                await this.checkPendingSync();
                await this.loadInitialData();
                this.showFeedback("✓ Jaringan terputus, tersimpan offline", false);
            }
        },

        showFeedback(msg, isError = false) {
            this.feedback.message = msg;
            this.feedback.isError = isError;
            this.feedback.show = true;
            setTimeout(() => {
                this.feedback.show = false;
            }, 2500);
        },

        async setBudget(amount) {
            if (this.isOffline) {
                this.showFeedback("❌ Harus online untuk mengubah anggaran", true);
                return;
            }
            
            try {
                const res = await fetch(`/api/v1/budget/${this.currentMonth}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ total_budget: parseInt(amount, 10) })
                });
                
                if (res.ok) {
                    await this.loadInitialData();
                    this.showFeedback("✓ Anggaran diperbarui", false);
                } else {
                    const err = await res.json();
                    this.showFeedback("Gagal mengatur anggaran: " + err.error, true);
                }
            } catch (err) {
                console.error('Set budget error:', err);
                this.showFeedback("❌ Kesalahan jaringan", true);
            }
        },

        exportMarkdown() {
            window.location.href = `/api/v1/export/markdown?year_month=${this.currentMonth}`;
        },

        updateOfflineStatus(status) {
            this.isOffline = status;
            console.log(`Application is now ${status ? 'Offline' : 'Online'}`);
            
            if (!status && this.pendingSync > 0) {
                console.log('Back online. Ready to sync pending transactions...');
                this.syncTransactions();
            }
        }
    });
});
