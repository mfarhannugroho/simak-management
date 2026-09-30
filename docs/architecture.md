# Arsitektur SIMAK PT INKA (Persero)

## Ringkasan Port & Service

| Service | Port | Keterangan |
|---|---|---|
| Frontend (Vue 3 + Vite) | `5173` | Antarmuka Pengguna SPA |
| Backend (Go + Gin) | `8080` | REST API Server & WebSocket |
| PostgreSQL 16 | `5432` | Relational Database & Migrations |
| Redis 7 | `6379` | Token Blacklist, Cache, Session |
| MinIO (S3 API) | `9000` | Object Storage Dokumen & Foto |
| MinIO Console | `9001` | Dashboard Administrasi MinIO |
| Swagger Docs | `8080/swagger/index.html` | Dokumentasi Interaktif API |

## Panduan Menjalankan

### Menggunakan Docker Compose (Direkomendasikan)
```bash
docker compose up -d
```

### Menjalankan Backend Secara Lokal
```bash
cd backend
go run cmd/server/main.go
```

### Menjalankan Frontend Secara Lokal
```bash
cd frontend
npm run dev
```

### Menjalankan Migrasi Database
```bash
cd backend
go run cmd/migrate/main.go up
```
