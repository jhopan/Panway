# AGENTS.md — Panway

Konteks kerja untuk AI agent dan manusia yang menggarap **Panway**.
Baca dulu sebelum mengubah kode. Perbarui file ini kalau arsitektur atau
aturannya berubah.

---

## 1. Apa itu Panway

Lapisan komersial di depan **PanRouter**.

```
Pelanggan ──► Panway ──► PanRouter ──► model AI
 (API key)   (akun, bayar, (routing, quota
              batas,usage)   subscription)
```

| Panway | PanRouter |
|---|---|
| Akun pelanggan, login, verifikasi | Routing model/channel |
| Penerbitan API key | Kuota per subscription |
| Saldo, top-up (PanWallet) | Rotasi akun multi |
| Metering pemakaian, log | Harga upstream |
| Rate limit per token | |
| Panel admin + pelanggan | |

Aturan: kalau fitur bisa regenerate dari PanRouter, jangan tulis ulang di
Panway. Satu sumber kebenaran.

## 2. Basis kode

Fork **[One API](https://github.com/songquanpeng/one-api)** (JustSong), lisensi
MIT. Modul: `github.com/jhopan/Panway`. Repo: `github.com/jhopan/Panway`,
branch `master`.

- Kredit JustSong **wajib** ada di README + footer web. Jangan dibuang.
- Fork lain (new-api/sub2api) tidak dipakai.

| Path | Isi |
|---|---|
| `common/config/config.go` | semua env var |
| `common/token-rate-limit.go` | limiter RPM/TPM/concurrency per token |
| `common/telegram.go` | client Bot API Telegram |
| `common/walletclient/client.go` | client HTTP PanWallet |
| `common/i18n/locales/` | `en.json`, `id.json` (pesan server) |
| `controller/pricing.go` | endpoint publik `/api/pricing` (katalog + harga) |
| `controller/pricing_admin.go` | endpoint admin edit harga per model |
| `pages/Console.css` | gaya bersama tabel/form console |
| `components/Sidebar.js` | navigasi console (rail kiri) |
| `components/ThemeToggle.js` | tombol light/dark/system |
| `helpers/theme.js` | simpan & terapkan mode tema |
| `semantic-dark.css` | override Semantic UI untuk mode gelap |
| `pages/Catalogue/` | halaman Models & Pricing |
| `components/Chrome.js` | memilih chrome per route |
| `pages/Landing/` | landing page marketing di `/` |
| `model/` | GORM model + auto-migrate |
| `controller/` | handler HTTP |
| `router/api.go` | daftar route `/api/*` |
| `relay/` | inti proxy + billing |
| `relay/billing/ratio/panrouter.go` | rasio harga per model PanRouter |
| `middleware/` | auth, rate limit, distributor, language |
| `web/default/` | frontend React (CRA) → `web/build/default/` |

Tema: **hanya `default`**. `web/THEMES` berisi satu baris. Jangan tambah tema
tanpa build-nya, karena server mencari `web/build/<tema>` saat start.

## 3. Menjalankan lokal

```bash
cd "C:/Users/ACER/Documents/Project/panway"
go run . --port 3110 --log-dir logs
```

Buka `http://127.0.0.1:3110`.

**Jebakan Windows (MSYS/git-bash):**
- `go run .` **wajib**. `panway.exe` → "Exec format error"; `cmd.exe /c
  panway.exe` → "Access is denied".
- Kalau "address already in use", ada proses lama masih pegang port:
  `netstat -ano | grep ":3110.*LISTENING"` → `cmd.exe /c "taskkill /PID
  <pid> /F"`.
- **Build frontend harus dengan server dimatikan.** Go meng-embed
  `web/build/*` ke binary, jadi While server jalan, `npm run build` gagal
  dengan `rm: cannot remove '../build/default': Device or resource busy`.
  Urutan yang benar: kill server → `npm run build` → `go run .`.
- `go build` ~1-2 menit. `npm install` ~5-6 menit.
- Argumen untuk program native (go, git, node) **tidak** diterjemahkan MSYS.
  Pakai `C:/Users/...`; `cd /c/Users/...` untuk perintah bash.

CLI setup klien (tanpa menyalakan server):

```bash
go run . connect --key sk-... --base-url http://127.0.0.1:3110/v1 --client all
# klien: all|env|curl|python|node|cursor|continue|aider
# --env-file my.env menulis export OPENAI_BASE_URL/OPENAI_API_KEY
```

Data lokal: SQLite di `data/panway.db` (`SQLITE_PATH`). Jangan dihapus.
Admin default: `root` / `123456` — **ganti sebelum produksi**.

## 4. Environment

Semua default ada di `common/config/config.go`. `.env` tidak di-commit;
`.env.example` yang jadi acuan.

| Env | Guna |
|---|---|
| `PORT` | port server (default 3110) |
| `SESSION_SECRET` | **wajib** random panjang di produksi |
| `SQLITE_PATH` | lokasi file SQLite |
| `PANWALLET_ENABLED` | nyalakan integrasi PanWallet |
| `PANWALLET_URL` / `PANWALLET_KEY` | alamat + product key PanWallet |
| `PANWALLET_QUOTA_PER_IDR` | konversi IDR → kuota |
| `PANWALLET_PUBLIC_URL` | link yang dilihat pelanggan |
| `TELEGRAM_VERIFICATION_ENABLED` | nyalakan verifikasi Telegram |
| `TELEGRAM_BOT_TOKEN` | token bot dari @BotFather |
| `TELEGRAM_BOT_USERNAME` | username bot tanpa "@" |
| `TELEGRAM_API_BASE_URL` | default `api.telegram.org` |
| `CORS_ALLOWED_ORIGINS` | origin browser yang boleh memanggil API (koma); kosong = same-origin |
| `TRUSTED_PROXIES` | alamat reverse proxy yang dipercaya (koma); kosong = tidak ada |
| `MAX_REQUEST_BODY_MB` | batas body request, termasuk hasil dekompresi gzip (default 32) |
| `SESSION_COOKIE_SECURE` | cookie sesi hanya lewat HTTPS |
| `HSTS_MAX_AGE` / `HSTS_INCLUDE_SUBDOMAINS` | HSTS; `0` = mati |
| `PRICING_CACHE_SECONDS` | TTL cache `/api/pricing` (default 30; `0` = mati) |
| `ESTIMATED_USAGE_MODE` | `local` (default) menormalkan usage estimated; `upstream` menagih angka upstream |
| `SQLITE_BUSY_TIMEOUT` | timeout lock SQLite (ms) |
| `SQL_MAX_OPEN_CONNS` / `SQL_MAX_IDLE_CONNS` | pool DB; default SQLite 1 |
| `SQLITE_BACKUP_ENABLED` / `SQLITE_BACKUP_DIR` | backup `VACUUM INTO` otomatis |
| `SQLITE_BACKUP_INTERVAL_HOURS` / `SQLITE_BACKUP_KEEP` | jadwal + jumlah snapshot |
| `LOG_RETENTION_DAYS` | hapus log lebih tua dari N hari; `0` = selamanya |
| `RETRY_TIMES` | retry ke channel lain saat upstream gagal |
| `RELAY_TIMEOUT` | timeout upstream (detik); `0` = tanpa batas |
| `TOKEN_ENCRYPTION_KEY` / `TOKEN_KEY_FILE` | kunci enkripsi API key; kosong = file kunci otomatis |
| `LOW_BALANCE_THRESHOLD` / `LOW_BALANCE_NOTIFY_INTERVAL_HOURS` | reminder Telegram saldo rendah |
| `ENABLE_METRICS` / `METRICS_TOKEN` | endpoint `/metrics` Prometheus |

## 5. Aturan kode

- Go 1.20, modul `github.com/jhopan/Panway`.
- **Tanpa stub.** Fungsi yang belum jelas harus lempar error atau dihapus.
- Cek `err` selalu. `logger.Error(ctx, ...)` untuk log,
  `abortWithMessage` untuk error ke user.
- `gofmt` sebelum commit. Import: standar → pihak ketiga → internal.
- **Never commit:** `.env`, `data/`, `logs/`, `*.exe`, `*.db`, `web/build/`.

### API response

Semua endpoint membalas:

```json
{ "success": true, "message": "", "data": {} }
```

Frontend (`web/default/src/helpers`) bergantung pada bentuk ini.

### Bahasa

Tiga aturan wajib:

1. **Default English.** Bahasa kedua: Indonesia (`id`). Mandarin dihapus.
2. **Tidak boleh ada teks Mandarin/Jepang/Korea di source.** Semua pesan
   user-facing ditulis English. Nama merek provider resmi (Baidu Qianfan,
   Alibaba Tongyi, iFlytek Spark, ByteDance Doubao, Zhipu ChatGLM) boleh tetap
   memakai nama aslinya.
3. Kalau menambah string user-facing baru di backend, tambahkan juga ke
   `common/i18n/locales/en.json` **dan** `id.json`.

Bila memang harus mencocokkan pesan Mandarin dari provider, pakai rune escape
(`"\u5df2\u6b20\u8d39"`) supaya sumber tetap bebas byte CJK — contoh ada di
`monitor/manage.go`.

## 6. Rate limit per token

Tiga batas, diatur di panel **Token → Edit**. `0` = tanpa batas (RPM jatuh ke
`GLOBAL_API_RATE_LIMIT`).

| Kolom | Arti | Implementasi |
|---|---|---|
| `rate_limit_rpm` | request per menit | sliding window, `common/token-rate-limit.go` |
| `rate_limit_tpm` | token per menit | sliding window, token **dari usage upstream** |
| `concurrency_limit` | request paralel | counter in-flight |

Dua hal yang mudah salah:

1. **TPM pakai angka token asli dari upstream, bukan estimasi lokal.**
   Estimasi lokal bisa jauh lebih kecil (misal 8 vs 497 token, karena upstream
   menambah system prompt sendiri). Estimasi lokal hanya buat estimasi kuota.
2. **Error rate limit tidak boleh di-retry ke channel lain.** Kalau di-retry,
   pesan Panway tertimpa pesan upstream. `controller/relay.go` mengecualikan
   `token_rate_limit_exceeded` dari `shouldRetry`.

## 6b. Tagihan stream & saldo token

Dua aturan yang menjaga tagihan tetap benar saat dijual. Jangan diubah tanpa
menguji ulang lewat `relay/adaptor/openai` dan `model/token.go`.

**1. Usage `estimated` dari upstream tidak ditagih apa adanya.** PanRouter
(sama seperti provider lain) tidak bisa mengukur token saat streaming dan
mengirim `"estimated": true` dengan prompt yang digelembungkan (uji 2026-11:
prompt kecil dilaporkan 2500, padahal non-stream 501). `StreamHandler` di
`relay/adaptor/openai/main.go` memilih usage exact bila ada (sebagian
upstream mengirim estimasi dulu lalu angka asli), dan
`NormalizeEstimatedUsage` di `relay/adaptor/openai/helper.go` mengganti usage
yang masih `estimated` dengan hitungan lokal: prompt dari request +
overhead prompt upstream yang sudah dipelajari, completion dari teks hasil
stream. Chunk usage yang dikirim ke pelanggan ikut ditulis ulang, jadi angka
yang dilihat pelanggan sama dengan yang ditagih. Usage exact dibiarkan apa
adanya.

Overhead prompt (selisih prompt upstream exact vs hitungan lokal, mis. 499−10)
dipelajari dari setiap usage exact (non-stream atau stream yang mengirim angka
asli) dan disimpan di memori per nama model. Konsekuensi: **stream pertama
untuk sebuah model setelah restart** belum punya overhead, jadi ditagih
dengan hitungan lokal saja (lebih murah, bukan lebih mahal); stream berikutnya
setelah ada usage exact sudah setara non-stream. `ESTIMATED_USAGE_MODE=upstream`
mematikan normalisasi dan menagih estimasi upstream apa adanya kalau operator
ingin menyamakan dengan meter upstream.

**2. Saldo token tidak pernah minus.** `preConsumeQuota`
(`relay/controller/helper.go`) selalu memanggang kuota token, termasuk saat
user dianggap "trusted" (saldo besar). `model.PreConsumeTokenQuota` memakai
satu `UPDATE ... WHERE remain_quota >= ?` (`reserveTokenQuota`) supaya dua
request paralel tidak bisa memakai kuota yang sama, dan `PostConsumeTokenQuota`
menagih paling banyak sisa saldo (`consumeRemainingTokenQuota`). Kalau
pre-consume gagal, cache kuota user dikembalikan dari DB.

Konsekuensi yang perlu diketahui:

- Setiap request relay menulis kuota token + user di DB (pre-consume dan
  post-consume). `BatchUpdateEnabled` tidak berlaku untuk mutasi saldo token,
  karena pengecekan dan pemotongan saldo harus atomik.
- Estimasi biaya pre-consume memakai `PreConsumedQuota + prompt lokal +
  max_tokens`; kalau biaya asli lebih besar, sisa saldo ditagih sampai 0 dan
  request berikutnya ditolak (`Insufficient token quota`).

## 7. Verifikasi Telegram

Telegram = **verifikasi**, bukan login.

Alur: daftar (email atau Google) → **User → Edit** → masukkan chat id (dari
`@userinfobot`) → **Send code** → bot kirim kode 6 digit → masukkan kode →
**Verify**.

- Kode terikat ke **user DAN chat**. Kode untuk satu chat tak bisa dipakai
  untuk chat lain.
- Berlaku 10 menit (`VerificationValidMinutes`).
- Endpoint: `GET /api/user/telegram/status`, `POST .../send`,
  `POST .../verify`, `POST .../unverify`.
- Kolom DB: `users.telegram_chat_id`, `users.telegram_verified`.

## 8. Login pelanggan

| Cara | Status | Setup |
|---|---|---|
| Google (OAuth) | sudah ada, lewat OIDC | `OidcClientId`, `OidcClientSecret`, `OidcWellKnown` di panel → OIDC, arahkan ke Google |
| Email + kode OTP | sudah ada | SMTP di panel → Email, nyalakan Email verification |
| Telegram | verifikasi saja (§7) | env `TELEGRAM_*` |

OIDC di One API sudah mendukung Google — tidak perlu kode tambahan.

## 9. Integrasi PanWallet

PanWallet memegang uang. Panway cuma mengubah IDR jadi kuota.

- Client: `common/walletclient/client.go`. Penghubung: `model/wallet_link.go`.
- Endpoint: `/api/user/wallet/link`, `/unlink`, `/sync`; admin
  `/api/user/wallet/list`, `/topup`.
- Alur: `hold` → `capture` (sukses) → `release` (gagal).
- **Belum aktif.** Butuh `PANWALLET_KEY` (product key `panway`).

## 10. PanRouter sebagai upstream

- Satu channel, tipe **OpenAICompatible**, `base_url` berakhiran `/v1`.
- Key PanRouter di tabel `channels` — **jangan** ditulis di repo.
- Respons non-stream dari PanRouter bisa punya `data: [DONE]` menempel di akhir
  body. `relay/adaptor/openai/json.go` (`DecodeJSONBody` /
  `StripStreamTerminator`) yang membersihkannya. **Jangan dihapus** — tanpa itu
  OpenAI SDK Python gagal parse ("Extra data").
- Harga per model: `relay/billing/ratio/panrouter.go`. Rasio dikalikan group
  ratio dan user multiplier → `quota`.
- Contoh model: `agnes/agnes-3.0-flash`, `th/deepseek-v4-flash:free`,
  `cf/@cf/qwen/qwen2.5-coder-32b-instruct`.

## 11. Test & verifikasi

Wajib **sebelum** bilang "selesai":

```bash
go build ./...        # hijau
go vet ./...          # bersih
go test ./...         # 7 paket ok
```

Frontend:

```bash
cd web/default && DISABLE_ESLINT_PLUGIN=true CI=false npm run build
```

Hasil build harus ada di `web/build/default/index.html`.

Lampirkan **bukti angka**: byte, HTTP status, jumlah baris, screenshot. Jangan
klaim tanpa bukti.

Uji manual yang sudah terbukti:

| Uji | Cara | Harapan |
|---|---|---|
| RPM | token `rpm=2`, 3 request | ke-3 → 429 |
| TPM | token `tpm=600`, 3 request | ke-3 → 429 |
| Concurrency | token `conc=2`, 4 paralel | 2 → 429 |
| Telegram | kirim kode → masukkan | `verified:true` di DB |
| SDK | OpenAI Python SDK | streaming & non-streaming jalan |

Cek bahasa sebelum commit:

```bash
python -c "
import os,re
pat=re.compile('[\u3040-\u30ff\u4e00-\u9fff\uac00-\ud7af]')
for root,dirs,fs in os.walk('.'):
    dirs[:]=[d for d in dirs if d not in ('.git','node_modules','data','logs','bin','build')]
    for f in fs:
        if f.endswith(('.go','.js','.jsx','.json','.css','.html')):
            p=os.path.join(root,f)
            if 'tencent' in p: continue
            n=len(pat.findall(open(p,encoding='utf-8',errors='replace').read()))
            if n: print(p,n)
"
```

Harus kosong. (`relay/adaptor/tencent/model.go` dikecualikan — 100% komentar,
string literal 0.)

## 12. Git

- Branch `master`, satu fitur satu commit, commit langsung ke `master`.
- **Setiap perubahan wajib di-commit dan di-push ke `origin/master` setelah
  verifikasi.** Jangan tinggalkan pekerjaan di working tree. Jika push gagal
  karena kredensial, laporkan dan jangan menimpa commit orang lain.
- Format: `feat:`, `fix:`, `docs:`, `ci:`, `refactor:`. Subjek ≤ 70 karakter,
  tanpa titik di akhir. Body menjelaskan **kenapa**.
- `commitlint` di CI **advisory** (tidak menggagalkan build).
- Author `jhopan`.
- **Jangan** `git push --force` ke branch yang sudah dipakai.
- Release: tag `v*` memicu GitHub Actions yang build. **Jangan** build
  release lokal, dan jangan push tag tanpa diminta.

## 12b. Tampilan: landing vs console

Tiga jenis chrome, dipilih di `components/Chrome.js`:

| Route | Chrome |
|---|---|
| `/` | Landing sendiri, tanpa Header global |
| `/models`, `/pricing` | Halaman marketing publik, tanpa sidebar (untuk pengunjung) |
| `/login`, `/register`, `/reset`, `/oauth/*` | Header lama, tanpa sidebar |
| `/connect` | Console: sidebar kiri + konten (wajib login) |
| `/console/models`, `/console/pricing` | Sidebar kiri + konten (wajib login) |
| sisanya (console) | Sidebar kiri + konten |

**Landing** (`pages/Landing/`):
- Menarik katalog & harga dari `GET /api/pricing` (publik, tanpa login).
- Hardcode nama model/price di frontend **dilarang** - harga harus dari API
  supaya panel admin tetap sumber kebenaran.
- Landing punya seksi janji layanan (`landing.promises.*`): integritas
  model, privasi prompt, tagihan yang bisa diaudit, dukungan. Jangan taruh
  setup koneksi di landing; itu tugas halaman console `/connect`.
- Header marketing selalu menautkan Models/Pricing ke halaman publik,
  termasuk saat pengguna sudah login. Hanya tombol Console yang masuk
  dashboard. Chip saldo (`landing.header.balance`) tetap ditampilkan untuk
  pengguna yang login.
- Dashboard console (`pages/Dashboard/`) dibuka dengan empat kartu cepat
  (`dashboard.quick.*`): saldo, jumlah API key, plan aktif, dan base URL
  connect — meniru struktur dashboard penyedia token, memakai data Panway.
- Halaman `/connect` (console, `pages/Connect/`) berisi langkah setup, snippet
  env/curl/Python/Node/CLI dengan tombol salin, daftar klien
  OpenAI-compatible, dan catatan keamanan. Halaman ini tidak memakai
  `MarketingShell`; akses lewat sidebar grup Situs atau kartu Connect di
  dashboard. Semua teks ditulis sendiri; jangan menyalin teks atau aset situs
  lain.
- i18n key di namespace `landing`, path `landing.how.steps.*` (bukan
  `landing.steps.*`).

**Sidebar** (`components/Sidebar.js`):
- Menu: Dashboard, Token, Top Up, Plans, Log, grup Katalog (Models, Pricing),
  grup Administrasi (Channel, Redemption, User), grup Situs (Connect,
  API & status, Security).
- Semua item sidebar menunjuk ke route console, **bukan** halaman marketing:
  Models → `/console/models`, Pricing → `/console/pricing`.
- Halaman katalog console memakai komponen yang sama dengan halaman publik
  lewat prop `embedded` (`pages/Catalogue/Models.js`,
  `pages/Catalogue/Pricing.js`). Jangan duplikasi isi katalog; tambah prop
  atau ekstrak komponen.
- Logo sidebar menuju `/dashboard` saat sudah login, `/` saat belum.
- Riwayat penggunaan ada di halaman Log (`/log`) — tabel `LogsTable` memuat
  token prompt/completion, kuota, model, token name, dan detail per request.
- Item aktif ditandai lewat `location.pathname`.
- Di bawah 900px jadi drawer dengan tombol toggle + overlay.
- `i18n.language` bisa berupa tag regional ("en-US"), jadi dropdown harus
  menormalkan ke "en"/"id" dulu, kalau tidak nilai dropdown kosong.

**Logo:** `web/default/public/logo-panway.svg` (huruf P, gradient indigo ->
cyan). `helpers/utils.js` `getLogo()` memakai path itu sebagai default.

## 12c. Tema gelap

Mode tema disimpan di `localStorage` key `panway_theme` (`light` | `dark` |
`system`) dan diterapkan sebagai class `light`/`dark` di `<html>`.
**Default-nya `dark`** (`DEFAULT_THEME_MODE` di `helpers/theme.js`).

Terapkan tema di dua tempat, jangan hanya satu:
- `index.js` memanggil `applyTheme()` sebelum render pertama.
- `public/index.html` punya inline script yang sama, karena CSS sudah mulai
  paints sebelum React mount. Tanpa itu halaman berkedip putih.

- `helpers/theme.js` — baca/simpan/apply, plus `toggleTheme()`.
  `system` mengikuti preferensi OS dan tetap ikut berubah selama app jalan.
- `index.js` memanggil `applyTheme()` **sebelum render pertama** supaya
  halaman tidak berkedip dengan background yang salah.
- Semua warna diambil dari token CSS `--pw-*` di `index.css`
  (`--pw-bg`, `--pw-surface`, `--pw-ink`, `--pw-muted`, `--pw-line`,
  `--pw-accent`, ...). Warna hex langsung di komponen = bug tema gelap.
- Semantic UI tidak punya dark theme, jadi `semantic-dark.css` berisi
  override `html.dark .ui.*` untuk card, tabel, tombol, label, menu,
  form, dan modal. Audit ulang override ini kalau ada komponen Semantic
  baru yang dipakai.
- Dark mode bukan sekadar membalik warna: tombol primary dibalik jadi
  terang di atas gelap, badge status memakai background tinted + teks
  terang, dan `--pw-faint` dinaikkan supaya label kecil tetap lolos
  kontras.

## 12d. Edit harga model

Admin mengubah harga dari halaman katalog, bukan dari JSON mentah di
Settings -> Operation -> Ratio.

| Endpoint | Akses | Body |
|---|---|---|
| `GET /api/pricing` | publik | - |
| `GET /api/pricing/all` | admin | - |
| `PUT /api/pricing` | admin | `{model_id, input_price, output_price}` |
| `DELETE /api/pricing` | admin | `{model_id}` |

Dua hal yang mudah salah:

1. **model_id dikirim di body, bukan di path.** Model id mengandung `/`
   (misal `agnes/agnes-3.0-flash`) sehingga tidak muat jadi path param Gin.
2. **Harga dalam rupiah per 1 juta token**, bukan satuan ratio internal.
   Backend yang mengonversi: `modelRatio = input_price / 1e6 / groupRatio`,
   `completionRatio = output_price / input_price`.

`GET /api/pricing/all` hanya mengembalikan model yang dilayani channel
aktif. Peta ratio juga memuat semua model default One API, dan itu bukan
yang bisa dilayani deployment ini.

`GET /api/pricing` juga menyertakan `requests_7d` dan `avg_latency_ms` dari
log 7 hari, dan setiap edit harga dicatat ke `price_histories`
(`GET /api/pricing/history`, jendela 72 jam) untuk sparkline di tabel harga.
Tabel harga menampilkan model, input/output, pengali, panggilan 7 hari,
latensi rata-rata, dan tren harga 72 jam.

## 12e. Keamanan produksi

Lapisan yang dipasang di `main.go` dan `middleware/` supaya instance aman
dijual. Semua default aman tanpa env tambahan; env hanya untuk membuka akses
yang memang dibutuhkan deployment.

| Middleware | Berkas | Perilaku |
|---|---|---|
| `SecurityHeaders` | `middleware/security-headers.go` | nosniff, frame DENY, referrer policy, permissions policy; HSTS hanya bila `HSTS_MAX_AGE` > 0 dan request HTTPS |
| `NoStoreAPI` | `middleware/security-headers.go` | `/api`, `/v1`, `/dashboard` → `Cache-Control: no-store` |
| `BodyLimit` | `middleware/body-limit.go` | 413 bila body > `MAX_REQUEST_BODY_MB`; `MaxBytesReader` juga membatasi body chunked |
| `CORS` | `middleware/cors.go` | hanya origin di `CORS_ALLOWED_ORIGINS`; kosong = tanpa header CORS |
| `Cache` | `middleware/cache.go` | cache 1 minggu hanya untuk aset ber-hash; SPA route `no-cache` |
| `GzipDecodeMiddleware` | `middleware/gzip.go` | body hasil dekompresi juga dibatasi `MAX_REQUEST_BODY_MB` |

Empat jebakan yang sudah diperbaiki, jangan dikembalikan:

1. **CORS `AllowAllOrigins` + `AllowCredentials`.** Itu membiarkan situs mana
   pun mengirim request ber-cookie ke console admin. Sekarang allowlist.
2. **`Cache-Control: max-age=604800` untuk semua path non-`/`.** SPA route
   (`/console/token`) menyajikan `index.html`; cache seminggu membuat app lama
   tetap jalan setelah deploy. Aset ber-hash tetap boleh lama.
3. **Gin mempercayai semua proxy.** Tanpa `SetTrustedProxies`, `ClientIP()`
   mengikuti `X-Forwarded-For` kiriman klien, jadi rate limit global dan cek
   subnet token bisa dilewati. Default sekarang tidak mempercayai proxy.
4. **Tidak ada batas body dan timeout header.** Satu request raksasa atau
   koneksi slowloris bisa menghabiskan memori/koneksi. `ReadHeaderTimeout`
   15 detik, `IdleTimeout` 120 detik, `MaxHeaderBytes` 1 MiB; `WriteTimeout`
   sengaja tidak diisi karena streaming.

Saat start, `SESSION_SECRET` kosong dengan `GIN_MODE=release` langsung gagal;
secret contoh atau < 32 karakter hanya jadi error log. Password root default
`123456` juga jadi error log (`model.WarnIfDefaultRootPassword`).

Perubahan admin (harga, plan, channel, opsi) dicatat sebagai log tipe
`Manage` lewat `controller/audit.go`; nilainya tidak ikut ditulis, hanya aksi
dan targetnya.

`/api/pricing` di-cache in-process (`PRICING_CACHE_SECONDS`, default 30 detik)
dan dibatalkan saat harga diedit lewat `PUT/DELETE /api/pricing` atau opsi
`ModelRatio` / `CompletionRatio` / `GroupRatio`. Response-nya juga dikirim
dengan `Cache-Control: public, max-age=<TTL>`.

## 12f. Operasional SQLite

SQLite dipakai sebagai database utama sampai trafik menuntut PostgreSQL.
`openSQLite` menyalakan WAL, `synchronous=NORMAL`, `busy_timeout`, dan
`foreign_keys`. Pool default SQLite **1 koneksi** (`SQL_MAX_OPEN_CONNS`),
karena SQLite hanya punya satu writer; pool besar justru memicu
`SQLITE_BUSY`. Pindah ke PostgreSQL/MySQL cukup isi `SQL_DSN`.

| Fitur | Berkas | Perilaku |
|---|---|---|
| Backup | `model/backup.go` | `VACUUM INTO` saat start lalu tiap `SQLITE_BACKUP_INTERVAL_HOURS`; simpan `SQLITE_BACKUP_KEEP` snapshot |
| Retensi log | `model/retention.go` | hapus log lama per 1000 baris, harian |
| Health | `controller/health.go` | `GET /healthz` → status, database, jumlah channel, versi; tanpa auth |
| Metrics | `controller/metrics.go` | `GET /metrics` (opt-in) counter request/error/token/quota + gauge channel |
| Shutdown | `main.go` | SIGINT/SIGTERM → `http.Server.Shutdown` 30 detik |
| Compose | `docker-compose.yml` | SQLite volume `/data`, nginx profile `tls`, Redis profile `redis` |

**Backup wajib menyertakan `TOKEN_KEY_FILE`** (default `data/token.key`).
Tanpa file itu, API key di backup tidak bisa didekripsi.

## 12g. Paket langganan

`model/plan.go`, `controller/plan.go`, halaman `/plans`.

- `plans` menyimpan nama, harga (`price_idr`), kuota, durasi, daftar model,
  status. Pembelian memotong saldo sebesar harga lalu menambah kuota.
- Validasi menolak `quota > price_idr`; tanpa aturan itu pembelian berulang
  bisa mencetak saldo.
- Pembelian menulis `users.plan_id` + `users.plan_expires_at` dan satu log
  tipe topup. Job harian membersihkan plan kedaluwarsa.
- `TokenAuth` menolak model di luar `plan.Models` selama plan aktif.
- Endpoint: `GET /api/plans` + `POST /api/plans/:id/buy` (user),
  `GET/POST/PUT/DELETE /api/plan` (admin).
- Riwayat penggunaan bisa diekspor CSV (`/api/log/self/export`, admin
  `/api/log/export`) dan setiap log topup punya kwitansi cetak
  `GET /api/log/receipt/:id` (pemilik atau admin).

## 12h. 2FA dan enkripsi API key

- **API key** disimpan sebagai ciphertext AES-GCM (`enc:v1:`) dengan lookup
  HMAC di `tokens.key_hash`. Kunci berasal dari `TOKEN_ENCRYPTION_KEY` atau
  file `TOKEN_KEY_FILE` 32 byte yang dibuat otomatis. `model.MigrateTokenKeys`
  mengenkripsi key lama saat start; console tetap menampilkan plaintext hasil
  dekripsi. Tanpa file/kunci yang sama, key tidak bisa dibuka.
- **2FA TOTP** (SHA-1, 6 digit, 30 detik) di `common/totp.go`. Setup/enable/
  disable di `/api/user/2fa*`; login password menghasilkan
  `require_2fa` lalu diselesaikan lewat `POST /api/user/login/2fa`.
  Login OAuth belum melewati langkah TOTP.
- Endpoint docs/status untuk pelanggan ada di `/docs` (console), membaca
  `/healthz`.

## 13. Yang belum dikerjakan

- Ganti password root (`123456` masih aktif) + `SESSION_SECRET` random.
- PanWallet belum aktif (butuh product key `panway`).
- Google login tinggal isi konfigurasi OIDC.
- Belum ada domain/deploy produksi; Dockerfile/compose belum diuji di mesin
  ini karena Docker tidak terpasang.
- Login OAuth belum meminta kode 2FA (baru login password).
- Endpoint Anthropic `/v1/messages` belum ada; saat ini hanya shape
  OpenAI-compatible (`/v1/chat/completions`, embeddings, images, audio).

## 14. Peringatan

- Jangan commit `.env`, `data/`, `logs/`, `*.exe`, `*.db`, `web/build/`.
- Jangan reboot mesin atau VPS tanpa diminta. VPS produksi: read-only SSH
  boleh, mutasi hanya atas perintah eksplisit.
- Jangan pakai `npm run build` untuk release produksi — CI yang bangun.
- Jangan ganti repo ini dengan fork lain; ini milik jhopan.
