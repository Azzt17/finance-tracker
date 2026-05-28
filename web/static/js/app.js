// Register Service Worker for PWA
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(reg => console.log('ServiceWorker registered with scope:', reg.scope))
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
        init() {
            // Register event listeners for network status on window
            window.addEventListener('online', () => this.updateOfflineStatus(false));
            window.addEventListener('offline', () => this.updateOfflineStatus(true));

            // Load initial app data
            this.loadInitialData();
        },

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

                this.categories = cats || [];
                this.budget = budg || { total_budget: 0, total_spent: 0, remaining_balance: 0 };
                this.savingsGoals = sav || [];
                this.recentTransactions = txs || [];
                
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
                // To be implemented: store in IndexedDB and sync later
                this.pendingSync++;
                this.showFeedback("Ditambahkan ke antrean offline", false);
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
                this.showFeedback("Kesalahan jaringan", true);
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
                alert("Anda harus online untuk mengubah anggaran.");
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
                    alert("Anggaran bulanan berhasil diperbarui.");
                } else {
                    const err = await res.json();
                    alert("Gagal mengatur anggaran: " + err.error);
                }
            } catch (err) {
                console.error('Set budget error:', err);
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
                // Implement trigger sync process here later
            }
        }
    });
});
