// Register Service Worker for PWA
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(reg => {
                reg.onupdatefound = () => {
                    const installingWorker = reg.installing;
                    if (installingWorker == null) return;
                    installingWorker.onstatechange = () => {
                        if (installingWorker.state === 'installed' && navigator.serviceWorker.controller) {
                            const financeStore = document.querySelector('[x-data]')?._x_dataStack?.[0]?.$store?.finance;
                            if (financeStore) financeStore.showFeedback("✨ Versi baru diunduh. Memuat ulang...", false);
                            setTimeout(() => { window.location.reload(); }, 2000);
                        }
                    };
                };
            })
            .catch(err => console.error('ServiceWorker registration failed:', err));
    });
}

document.addEventListener('alpine:init', () => {
    Alpine.store('finance', {
        currentView: 'input',
        currentMonth: new Date().toISOString().slice(0, 7),
        
        budget: { total_budget: 0, total_spent: 0, remaining_balance: 0 },
        savingsGoals: [],
        recentTransactions: [],
        categories: [],
        
        isLoading: false,
        isOffline: !navigator.onLine,
        pendingSync: 0,
        feedback: { show: false, message: '', isError: false },
        db: null,

        get quickAddCategories() {
            return this.categories
                .filter(c => c.is_quick_add === true || c.is_quick_add === 1)
                .sort((a, b) => a.sort_order - b.sort_order);
        },

        async init() {
            window.addEventListener('online', () => this.updateOfflineStatus(false));
            window.addEventListener('offline', () => this.updateOfflineStatus(true));
            await this.initDB();
            this.loadInitialData();
        },

        async initDB() {
            return new Promise((resolve, reject) => {
                const request = indexedDB.open('finance_db', 1);
                request.onupgradeneeded = (e) => {
                    const db = e.target.result;
                    if (!db.objectStoreNames.contains('sync_queue')) {
                        db.createObjectStore('sync_queue', { keyPath: 'client_transaction_id' });
                    }
                };
                request.onsuccess = (e) => {
                    this.db = e.target.result;
                    this.checkPendingSync();
                    resolve(this.db);
                };
                request.onerror = (e) => reject(e.target.error);
            });
        },

        async saveToSyncQueue(transaction) {
            return new Promise((resolve, reject) => {
                if (!this.db) return reject('DB not initialized');
                const tx = this.db.transaction('sync_queue', 'readwrite');
                const store = tx.objectStore('sync_queue');
                const reqGet = store.get(transaction.client_transaction_id);
                reqGet.onsuccess = () => {
                    const existing = reqGet.result || {};
                    const merged = { ...existing, ...transaction };
                    const reqPut = store.put(merged);
                    reqPut.onsuccess = () => resolve();
                    reqPut.onerror = () => reject(reqPut.error);
                };
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
                this.showFeedback("Mulai sinkronisasi...", false);
                const res = await fetch('/api/v1/sync/transactions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ transactions: queue })
                });

                if (res.ok) {
                    await this.clearSyncQueue();
                    this.pendingSync = 0;
                    await this.loadInitialData();
                    this.showFeedback(`✓ Sinkronisasi selesai`, false);
                } else {
                    this.showFeedback("❌ Gagal sinkronisasi", true);
                }
            } catch (err) {
                console.error("Sinkronisasi tertunda", err);
            }
        },

        changeView(viewName) {
            this.currentView = viewName;
            window.scrollTo(0, 0);
        },

        getCategoryName(id) {
            if (!id) return "Umum";
            const cat = this.categories.find(c => c.id === id);
            return cat ? `${cat.icon_emoji || ''} ${cat.name}`.trim() : "Unknown";
        },

        async loadInitialData() {
            this.isLoading = true;
            try {
                const month = this.currentMonth;
                const [cats, budg, sav, txs] = await Promise.all([
                    fetch('/api/v1/categories').then(r => r.ok ? r.json() : []),
                    fetch(`/api/v1/budget/${month}`).then(r => r.ok ? r.json() : { total_budget: 0, total_spent: 0, remaining_balance: 0 }),
                    fetch(`/api/v1/savings?year_month=${month}`).then(r => r.ok ? r.json() : []),
                    fetch(`/api/v1/transactions?year_month=${month}`).then(r => r.ok ? r.json() : [])
                ]);

                const pendingTxs = await this.getSyncQueue();
                this.categories = cats || [];
                this.savingsGoals = sav || [];
                
                const activePendingTxs = pendingTxs.filter(t => !t.is_deleted).map(t => ({...t, id: 'temp-'+t.client_transaction_id}));
                const pendingDeletes = pendingTxs.filter(t => t.is_deleted).map(t => t.client_transaction_id);
                const filteredServerTxs = (txs || []).filter(t => !pendingDeletes.includes(t.client_transaction_id));
                
                this.recentTransactions = [...activePendingTxs, ...filteredServerTxs].sort((a, b) => new Date(b.transacted_at) - new Date(a.transacted_at));
                
                let pendingSpent = activePendingTxs.reduce((sum, tx) => sum + tx.amount, 0);
                let serverPendingSpent = filteredServerTxs.filter(t => pendingTxs.find(p => p.client_transaction_id === t.client_transaction_id && p.is_deleted)).reduce((sum, tx) => sum + tx.amount, 0);
                let actualBudget = budg || { total_budget: 0, total_spent: 0, remaining_balance: 0 };
                
                this.budget = {
                    total_budget: actualBudget.total_budget,
                    total_spent: actualBudget.total_spent + pendingSpent - serverPendingSpent,
                    remaining_balance: actualBudget.remaining_balance - pendingSpent + serverPendingSpent
                };
                
            } catch (error) {
                console.error('Data load failed', error);
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
                await this.saveToSyncQueue(payload);
                await this.checkPendingSync();
                await this.loadInitialData();
                this.showFeedback("✓ Tersimpan di antrean", false);
                return;
            }

            try {
                const res = await fetch('/api/v1/transactions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                if (res.ok) {
                    await this.loadInitialData();
                    this.showFeedback("✓ Transaksi Tercatat", false);
                } else {
                    this.showFeedback("Gagal mencatat", true);
                }
            } catch (err) {
                await this.saveToSyncQueue(payload);
                await this.checkPendingSync();
                await this.loadInitialData();
                this.showFeedback("✓ Tersimpan di antrean", false);
            }
        },

        async deleteTransaction(clientTxId, id) {
            if (!confirm("Hapus transaksi ini?")) return;
            
            if (String(id).startsWith('temp-') || this.isOffline) {
                await this.saveToSyncQueue({ client_transaction_id: clientTxId, is_deleted: true });
                await this.checkPendingSync();
                this.showFeedback("✓ Dihapus (Offline)", false);
                await this.loadInitialData();
                return;
            }

            try {
                const res = await fetch(`/api/v1/transactions/${id}`, { method: 'DELETE' });
                if (res.ok) {
                    this.showFeedback("✓ Dihapus", false);
                    await this.loadInitialData();
                } else {
                    this.showFeedback("❌ Gagal dihapus", true);
                }
            } catch (e) {
                await this.saveToSyncQueue({ client_transaction_id: clientTxId, is_deleted: true });
                await this.checkPendingSync();
                this.showFeedback("✓ Antrean hapus", false);
                await this.loadInitialData();
            }
        },

        async setBudget(amount) {
            if (this.isOffline) return this.showFeedback("❌ Harus online", true);
            try {
                const res = await fetch(`/api/v1/budget/${this.currentMonth}`, {
                    method: 'PUT', headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ total_budget: parseInt(amount, 10) })
                });
                if (res.ok) {
                    await this.loadInitialData();
                    this.showFeedback("✓ Anggaran diperbarui", false);
                } else {
                    this.showFeedback("Gagal", true);
                }
            } catch (err) {
                this.showFeedback("❌ Kesalahan jaringan", true);
            }
        },

        async addCategory(name, icon, isQuickAdd) {
            if(this.isOffline) return this.showFeedback("Hanya bisa saat online", true);
            const payload = { name, icon_emoji: icon, is_quick_add: isQuickAdd, sort_order: this.categories.length + 1 };
            const res = await fetch('/api/v1/categories', { method: 'POST', body: JSON.stringify(payload) });
            if(res.ok) { this.showFeedback("✓ Kategori ditambah"); this.loadInitialData(); }
        },
        async deleteCategory(id) {
            if(this.isOffline) return this.showFeedback("Hanya bisa saat online", true);
            if(!confirm("Hapus kategori ini?")) return;
            const res = await fetch(`/api/v1/categories/${id}`, { method: 'DELETE' });
            if(res.ok) { this.showFeedback("✓ Dihapus"); this.loadInitialData(); }
        },
        
        async addSavings(name, target) {
            if(this.isOffline) return this.showFeedback("Hanya bisa saat online", true);
            const payload = { name, target_amount: parseInt(target), current_saved: 0, year_month: this.currentMonth, is_achieved: false };
            const res = await fetch('/api/v1/savings', { method: 'POST', body: JSON.stringify(payload) });
            if(res.ok) { this.showFeedback("✓ Tabungan ditambah"); this.loadInitialData(); }
        },
        async deleteSavings(id) {
            if(this.isOffline) return this.showFeedback("Hanya bisa saat online", true);
            if(!confirm("Hapus target tabungan?")) return;
            const res = await fetch(`/api/v1/savings/${id}`, { method: 'DELETE' });
            if(res.ok) { this.showFeedback("✓ Dihapus"); this.loadInitialData(); }
        },

        showFeedback(msg, isError = false) {
            this.feedback.message = msg;
            this.feedback.isError = isError;
            this.feedback.show = true;
            setTimeout(() => { this.feedback.show = false; }, 2500);
        },
        exportMarkdown() {
            window.location.href = `/api/v1/export/markdown?year_month=${this.currentMonth}`;
        },
        updateOfflineStatus(status) {
            this.isOffline = status;
            if (!status && this.pendingSync > 0) this.syncTransactions();
        }
    });
});
