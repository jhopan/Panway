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
| `components/Sidebar.js` | navigasi console (rail kiri) |
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
| `/` | Landing sendiri (header + footer marketing), tanpa Header global |
| `/login`, `/register`, `/reset`, `/oauth/*` | Header lama, tanpa sidebar |
| sisanya (console) | Sidebar kiri + konten |

**Landing** (`pages/Landing/`):
- Menarik katalog & harga dari `GET /api/pricing` (publik, tanpa login).
- Hardcode nama model/price di frontend **dilarang** - harga harus dari API
  supaya panel admin tetap sumber kebenaran.
- i18n key di namespace `landing`, path `landing.how.steps.*` (bukan
  `landing.steps.*`).

**Sidebar** (`components/Sidebar.js`):
- Menu: Dashboard, Token, Top Up, Log, Settings + grup Administrasi
  (Channel, Redemption, User) + grup Situs (About).
- Item aktif ditandai lewat `location.pathname`.
- Di bawah 900px jadi drawer dengan tombol toggle + overlay.
- `i18n.language` bisa berupa tag regional ("en-US"), jadi dropdown harus
  menormalkan ke "en"/"id" dulu, kalau tidak nilai dropdown kosong.

**Logo:** `web/default/public/logo-panway.svg` (huruf P, gradient indigo ->
cyan). `helpers/utils.js` `getLogo()` memakai path itu sebagai default.

## 13. Yang belum dikerjakan

- Ganti password root (`123456` masih aktif) + `SESSION_SECRET` random.
- PanWallet belum aktif (butuh product key `panway`).
- Google login tinggal isi konfigurasi OIDC.
- Belum ada domain/deploy produksi.
- Belum ada plan/subscription terpisah — semua pakai group `default`.

## 14. Peringatan

- Jangan commit `.env`, `data/`, `logs/`, `*.exe`, `*.db`, `web/build/`.
- Jangan reboot mesin atau VPS tanpa diminta. VPS produksi: read-only SSH
  boleh, mutasi hanya atas perintah eksplisit.
- Jangan pakai `npm run build` untuk release produksi — CI yang bangun.
- Jangan ganti repo ini dengan fork lain; ini milik jhopan.
