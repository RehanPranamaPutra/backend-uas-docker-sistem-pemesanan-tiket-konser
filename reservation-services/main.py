from fastapi import FastAPI, HTTPException
import redis
import os
import requests
import threading
import time

app = FastAPI()

# --- KONFIGURASI KONEKSI ---
REDIS_HOST = os.getenv("REDIS_HOST", "db-report")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))

# decode_responses=True sangat penting agar data yang diambil otomatis jadi String
r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

# --- FUNGSI AKTIVASI NOTIFIKASI REDIS ---
def enable_redis_notifications():
    """Mengaktifkan fitur 'Expired Notification' di Redis secara otomatis saat startup"""
    try:
        r.config_set('notify-keyspace-events', 'Ex')
        print("✅ Redis Keyspace Notifications (Expired) diaktifkan.")
    except Exception as e:
        print(f"❌ Gagal mengaktifkan konfigurasi Redis: {e}")

# --- WORKER: MENGEMBALIKAN STOK SAAT TIMEOUT ---
def redis_event_listener():
    """Mendengarkan event expired dari Redis untuk mengembalikan stok"""
    pubsub = r.pubsub()
    # Berlangganan ke channel khusus event expired database 0
    pubsub.subscribe("__keyevent@0__:expired")
    
    print("Background Worker: Menunggu tiket expired (Payment Timeout)...")
    
    for message in pubsub.listen():
        if message['type'] == 'message':
            expired_key = message['data']
            
            # Format kunci kita: lock:event_id:qty:user_id
            if expired_key.startswith("lock:"):
                parts = expired_key.split(":")
                if len(parts) == 4:
                    event_id = parts[1]
                    qty = int(parts[2])
                    
                    # KEMBALIKAN STOK KE REDIS
                    r.incrby(f"stock:{event_id}", qty)
                    print(f"⏰ [TIMEOUT] User telat bayar. Stok Konser {event_id} ditambah kembali {qty}.")

# Jalankan aktivasi konfigurasi
enable_redis_notifications()
# Jalankan listener di thread berbeda agar API tetap bisa melayani request
threading.Thread(target=redis_event_listener, daemon=True).start()


@app.get("/")
def root():
    return {"status": "running", "service": "Reservation Service (Python/FastAPI)"}

# --- ENDPOINT 1: RESERVE STOK (DIPANGGIL GO) ---
@app.post("/reserve/{event_id}/{qty}/{user_id}")
def reserve_stock(event_id: str, qty: int, user_id: str):
    stock_key = f"stock:{event_id}"
    # Kunci unik: mengandung qty agar saat expired worker tahu berapa yang harus dikembalikan
    lock_key = f"lock:{event_id}:{qty}:{user_id}"
    
    # 1. CEK DOUBLE BOOKING (Mencegah user yang sama klik berkali-kali)
    if r.exists(lock_key):
        raise HTTPException(status_code=400, detail="Anda sudah memiliki pesanan aktif, silakan bayar dulu")

    # 2. AMBIL STOK DARI REDIS
    current_stock = r.get(stock_key)
    
    # 3. AUTO-SYNC: JIKA REDIS KOSONG, AMBIL DARI LARAVEL
    if current_stock is None:
        print(f"Stok {event_id} kosong di Redis. Mengambil dari Laravel...")
        try:
            laravel_url = f"http://catalog-service:8000/api/concerts/{event_id}"
            resp = requests.get(laravel_url, timeout=5)
            if resp.status_code == 200:
                initial_val = resp.json()['stock']
                r.set(stock_key, initial_val)
                current_stock = initial_val
            else:
                raise HTTPException(status_code=404, detail="Konser tidak terdaftar di Katalog")
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Koneksi ke Laravel Gagal: {str(e)}")

    # 4. VALIDASI KUOTA
    if int(current_stock) < qty:
        raise HTTPException(status_code=400, detail="Maaf, stok tiket tidak mencukupi")

    # 5. EKSEKUSI: POTONG STOK & PASANG GEMBOK (TTL 300 detik = 5 Menit)
    r.decrby(stock_key, qty)
    r.setex(lock_key, 300, "pending")
    
    print(f"🎫 [RESERVE] User {user_id} memesan {qty} tiket Konser {event_id}.")
    return {"status": "success", "message": "Stok berhasil dikunci selama 5 menit"}

# --- ENDPOINT 2: KONFIRMASI BAYAR (DIPANGGIL GO) ---
# Pastikan urutan parameter di Python seperti ini:
@app.post("/confirm-payment/{event_id}/{qty}/{user_id}")
def confirm_payment(event_id: str, qty: int, user_id: str):
    # Kunci harus sama persis formatnya dengan saat /reserve
    lock_key = f"lock:{event_id}:{qty}:{user_id}"
    
    if r.exists(lock_key):
        r.delete(lock_key)
        print(f"💰 PAID: {lock_key} dihapus.")
        return {"status": "success"}
    
    # Jika tidak ketemu, berarti sudah expired (timeout)
    raise HTTPException(status_code=408, detail="Gembok sudah kedaluwarsa")