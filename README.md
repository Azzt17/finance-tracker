# 💰 Finance Tracker

Sebuah aplikasi pelacak keuangan pribadi (Personal Finance Tracker) berarsitektur *Offline-First Progressive Web App* (PWA). Aplikasi ini dirancang untuk pencatatan keuangan harian yang cepat, ringan, dan dapat di-host sendiri (*self-hosted*).

## ✨ Fitur Utama

- **Offline-First PWA:** Dapat diinstal di *smartphone* (Add to Home Screen), bisa dibuka tanpa internet, dan akan melakukan sinkronisasi latar belakang saat koneksi kembali (*Pending Sync*).
- **Quick Add Transactions:** Pencatatan pengeluaran super cepat dengan antarmuka angka yang besar dan dukungan kategori berbasis Emoji.
- **Budgeting & Savings Goals:** Pantau sisa anggaran bulanan dan kelola progres target tabungan Anda.
- **Export Data:** Ekspor riwayat transaksi Anda kapan saja dalam format Markdown.
- **Single Binary / Embedded Assets:** Frontend (HTML, CSS, JS) di-embed langsung ke dalam *binary* Go, membuat proses deployment sangat ringkas.
- **Basic Authentication:** Keamanan bawaan untuk melindungi akses data finansial Anda saat di-deploy ke internet.

## 🛠️ Teknologi yang Digunakan

- **Backend:** [Go](https://golang.org/) (Golang) dengan standard library `net/http`.
- **Database:** [SQLite](https://sqlite.org/index.html) - ringan dan tidak butuh server database terpisah.
- **Frontend:** HTML5, [Pico CSS](https://picocss.com/) (untuk styling yang bersih & dark mode), dan [Alpine.js](https://alpinejs.dev/) (untuk reaktivitas & state management).

## 🚀 Panduan Pengembangan Lokal (Local Development)

### Prasyarat
- Go 1.21 atau lebih baru.

### Cara Menjalankan
1. Clone repositori ini.
2. Jalankan perintah berikut untuk memastikan semua *test* berjalan:
   ```sh
   go test ./...
   ```
3. Jalankan server:
   ```sh
   go run ./cmd/server
   ```
4. Buka browser dan akses `http://localhost:8080`. Data lokal akan disimpan otomatis dalam file `finance-tracker.db`.

## ⚙️ Konfigurasi (Environment Variables)

Aplikasi ini dikonfigurasi menggunakan variabel lingkungan (*Environment Variables*):

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `ADDR` | Alamat/Port *listen* HTTP untuk aplikasi. | `:8080` |
| `DATABASE_URL` | Koneksi SQLite (*Data Source Name*). | `file:finance-tracker.db?...` |
| `APP_USERNAME` | Username untuk Basic Auth. Aktif jika password juga diisi. | *(Kosong = Dinonaktifkan)* |
| `APP_PASSWORD` | Password untuk Basic Auth. Aktif jika username juga diisi. | *(Kosong = Dinonaktifkan)* |
| `CORS_ALLOWED_ORIGIN`| Batasan domain untuk akses API secara *cross-origin*. | *(Kosong = Same-origin)* |

## 🌐 Deployment (Cloudflare Tunnel & Docker)

Untuk penggunaan pribadi yang aman dan mudah diakses dari internet publik, sangat disarankan menggunakan **Docker Compose** dan **Cloudflare Tunnel**.

1. Masuk ke direktori deploy:
   ```sh
   cd deploy/cloudflare
   ```
2. Salin file contoh *environment*:
   ```sh
   cp .env.example .env
   ```
3. Edit file `.env`. Pastikan untuk mengatur `APP_USERNAME` dan `APP_PASSWORD` agar aplikasi Anda aman.
4. Jalankan *container*:
   ```sh
   docker compose up -d --build
   ```

*(Lihat `deploy/cloudflare/README.md` untuk detail instruksi setup Cloudflare).*
