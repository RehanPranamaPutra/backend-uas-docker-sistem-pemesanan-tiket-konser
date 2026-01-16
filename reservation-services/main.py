from fastapi import FastAPI, HTTPException
import redis
import os
import requests

app = FastAPI()

# --- KONEKSI REDIS ---
REDIS_HOST = os.getenv("REDIS_HOST", "db-report")
REDIS_PORT = os.getenv("REDIS_PORT", 6379)
r = redis.Redis(host=REDIS_HOST, port=int(REDIS_PORT), decode_responses=True)

@app.get("/")
def root():
    return {"message": "Ticket Reservation Service is Running on Port 5002"}

@app.post("/reserve/{event_id}/{qty}/{user_id}")
def reserve_stock(event_id: str, qty: int, user_id: str):
    stock_key = f"stock:{event_id}"
    lock_key = f"lock:{event_id}:{user_id}"
    
    current_stock = r.get(stock_key)
    
    if current_stock is None:
        # --- AUTO-SYNC DARI LARAVEL ---
        # Gunakan port 8000 (port asli Laravel di dalam container)
        laravel_url = f"http://catalog-service:8000/api/concerts/{event_id}"
        try:
            response = requests.get(laravel_url, timeout=5)
            if response.status_code == 200:
                data = response.json()
                initial_stock = data['stock'] 
                r.set(stock_key, initial_stock)
                current_stock = initial_stock
            else:
                raise HTTPException(status_code=404, detail="Konser tidak ada di Laravel")
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"Koneksi ke Laravel gagal: {str(e)}")

    if int(current_stock) < qty:
        raise HTTPException(status_code=400, detail="Stok tidak mencukupi")

    # Proses Pengurangan & Lock
    new_stock = r.decrby(stock_key, qty)
    r.setex(lock_key, 60, qty) # Simpan qty agar bisa dikembalikan jika timeout

    return {"status": "success", "remaining_stock": new_stock}

@app.post("/confirm-payment/{event_id}/{user_id}")
def confirm_payment(event_id: str, user_id: str):
    lock_key = f"lock:{event_id}:{user_id}"
    if r.exists(lock_key):
        r.delete(lock_key)
        return {"status": "success", "message": "Gembok dilepas"}
    return {"status": "error", "message": "Gembok tidak ditemukan"}