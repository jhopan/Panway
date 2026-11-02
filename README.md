# Panway

**AI API gateway untuk menjual token AI.**

Panway berdiri di depan PanRouter. Satu channel OpenAI-compatible mengarah ke
PanRouter; Panway yang mengurus sisanya: pengguna, API key, kuota, harga
per model, pembayaran, dan log pemakaian.

```
Pelanggan ──Bearer <token Panway>──▶ Panway ──▶ PanRouter ──▶ provider AI
                                        │
                                        └──▶ PanWallet (saldo IDR, QRIS)
```

## Yang Panway urus

| | |
|---|---|
| **Pengguna** | Registrasi, login, saldo (`quota`) |
| **API key** | Token milik pelanggan: kuota, masa berlaku, daftar model, IP whitelist |
| **Harga** | Rasio harga per model (`ModelRatio` + `CompletionRatio`), dikali `groupRatio` |
| **Pembayaran** | Top up via PanWallet (IDR), atau redeem code |
| **Pemakaian** | Log per request: token, biaya, model, channel |
| **Trafik** | Retry, load balancing antar channel, priority, affinity |
| **Model** | `ModelMapping` — bisa menjual model dengan nama sendiri |

## Kebutuhan

- Go 1.20+ (dibangun dengan Go 1.26)
- Node 18+ hanya untuk membangun frontend
- SQLite (default), PostgreSQL, atau MySQL

## Menjalankan

```bash
cp .env.example .env
# isi .env

# build frontend (sekali)
cd web/default && npm install && npm run build && cd ../..

# build & jalankan
go build -o panway .
./panway --port 3000
```

Buka `http://localhost:3000`. Akun root pertama dibuat otomatis:
`root` / `123456` — **ganti passwordnya sebelum dipakai produksi**.

## Menyambungkan PanRouter

Di panel admin → **渠道管理 (Channel Management)** → tambah channel:

| Field | Nilai |
|---|---|
| 类型 (Type) | `OpenAI Compatible` (nilai `50`) |
| 密钥 (Key) | API key PanRouter-mu |
| 代理 (Base URL) | `http://<panrouter>:20128/v1` — **harus berakhiran `/v1`** |
| 模型 (Models) | Daftar model yang dijual, dipisah koma |
| 分组 (Group) | `default` |

Base URL wajib berisi `/v1`: untuk tipe `OpenAI Compatible`, one-api menyusun
URL dengan `baseURL + path tanpa /v1`, jadi `baseURL` harus sudah `/v1`.

### Model mapping (nama model sendiri)

Field **模型重定向 (Model Mapping)** pada channel memetakan nama yang dilihat
pelanggan ke nama model asli di PanRouter:

```json
{ "jago-1": "agnes/agnes-3.0-flash", "senior-code": "cf/@cf/qwen/qwen2.5-coder-32b-instruct" }
```

Pelanggan memanggil `jago-1`; PanRouter menerima `agnes/agnes-3.0-flash`.

### Harga per model

Panel admin → **设置 (Settings) → 模型倍率 (Model Ratio)**. Format:

```json
{
  "agnes/agnes-3.0-flash": 0.05,
  "th/deepseek-v4-flash:free": 0.07
}
```

`CompletionRatio` menentukan rasio harga output terhadap input.

Rumus biaya:

```
biaya = ceil((prompt_token + completion_token × completion_ratio) × model_ratio × group_ratio)
```

Nilai default untuk model PanRouter ada di `relay/billing/ratio/panrouter.go`.
Override dari panel admin selalu menang.

## Integrasi PanWallet

PanWallet adalah pemilik uang. Panway hanya mencerminkan saldo menjadi kuota.

```bash
PANWALLET_ENABLED=true
PANWALLET_URL=https://panwallet.jhopan.my.id
PANWALLET_KEY=<product key untuk produk "panway">
PANWALLET_QUOTA_PER_IDR=1        # 1 kuota = Rp1
```

### Alur

1. Pelanggan membuat wallet di PanWallet (dapat ID pendek 8 karakter).
2. Di panel Panway → **Top Up**, pelanggan memasukkan ID wallet-nya.
   Panway memverifikasi ke PanWallet sebelum menyimpan.
3. Pelanggan membayar invoice QRIS di PanWallet.
4. Pelanggan menekan **Sinkronkan** di halaman Top Up. Panway membaca saldo
   PanWallet secara live, memastikan cukup, lalu menambah kuota.

API:

| Method | Path | Auth |
|---|---|---|
| POST | `/api/user/wallet/link` | user — body `{"public_id":"ABC12345"}` |
| POST | `/api/user/wallet/unlink` | user |
| GET | `/api/user/wallet` | user — wallet + saldo live |
| POST | `/api/user/wallet/sync` | user — body `{"amount_idr":25000}` |
| GET | `/api/user/wallet/list` | admin |
| POST | `/api/user/wallet/topup` | admin — body `{"user_id":1,"amount_idr":25000}` |

## Endpoint API pelanggan

OpenAI-compatible:

```
POST /v1/chat/completions
POST /v1/messages
GET  /v1/models
```

## Perbedaan dari One API

Panway adalah fork One API (MIT) dengan perubahan:

- **Rebranding** ke "Panway" (nama sistem, teks antarmuka, module path, nama
  file database).
- **Harga default model PanRouter** (`relay/billing/ratio/panrouter.go`).
- **Integrasi PanWallet**: tautakan wallet, cek saldo live, sinkronisasi top-up
  (`common/walletclient/`, `controller/wallet.go`, `model/wallet_link.go`).
- **Parsing JSON toleran** untuk upstream yang menempelkan `data: [DONE]` ke
  respons non-streaming (`relay/adaptor/openai/json.go`).
- **Alamat SMTP aman IPv6** (`common/message/email.go`).

## Kredit & kontributor

| Peran | Nama |
|---|---|
| **Panway — maintainer, kontributor** | [jhopan](https://github.com/jhopan) |
| **One API — proyek asal** | [JustSong](https://github.com/songquanpeng) |

Panway adalah fork rebranded dari **One API** © JustSong, lisensi MIT. Kredit
asli dipertahankan di footer antarmuka, `LICENSE`, dan riwayat commit file
tersentuh — sesuai syarat lisensi MIT.

Kontribusi Panway di atas fork upstream:

- Rebranding penuh ke "Panway" (nama sistem, teks antarmuka dua bahasa, module
  path, nama database, nama file log, binary, image Docker).
- Integrasi **PanWallet**: tautakan wallet, cek saldo live, sinkronisasi top-up
  IDR ke kuota.
- Harga default model **PanRouter**.
- Parsing JSON toleran untuk upstream yang menempelkan `data: [DONE]` ke respons
  non-streaming.
- Pengiriman `User-Agent` browser saat ambil gambar dari URL user.
- Alamat SMTP aman IPv6.

## Lisensi

MIT. Berasal dari [One API](https://github.com/songquanpeng/one-api) © JustSong.
Copyright notice asli dipertahankan sesuai syarat lisensi. Lihat `LICENSE`.

## Peringatan

Menjual ulang akses ke akun langganan AI dapat melanggar ketentuan layanan
provider. Risiko itu sepenuhnya ada di tangan operator. Proyek ini hanya
menyediakan lapisan teknis.
