FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

#cpopy semua file kedalam container
COPY . .

EXPOSE 5002

#jalankan aplikasi
CMD ["uvicorn","main:app", "--host","0.0.0.0","--port","5002"]