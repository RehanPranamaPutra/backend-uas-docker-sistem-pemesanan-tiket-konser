# Gunakan image golang resmi
FROM golang:1.20-alpine

# Set folder kerja di dalam container
WORKDIR /app

# Copy file library dulu agar proses build lebih cepat
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode program
COPY . .

# Build program Go menjadi binary bernama 'main'
RUN go build -o main .

# Order service berjalan di port 8080
EXPOSE 8080

# Jalankan aplikasinya
CMD ["./main"]