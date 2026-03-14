# POS App Backend — Go (Gin + GORM + PostgreSQL + Redis)

Production-grade REST API backend for the **POS Flutter application**.

---

## 🏗 Architecture

```
pos_backend/
├── cmd/server/         → main.go (entrypoint, routes, DI)
├── config/             → Environment config loader
├── internal/
│   ├── auth/           → JWT token generation & validation
│   ├── handlers/       → HTTP request handlers (controllers)
│   ├── middleware/      → Auth, CORS, logging, rate limiting
│   ├── models/         → GORM domain models
│   └── services/       → Business logic layer
└── pkg/
    └── database/       → PostgreSQL + Redis connection
```

## 🚀 Quick Start

### 1. Prerequisites
- Go 1.22+
- PostgreSQL 15+
- Redis 7+

### 2. Setup
```bash
cp .env.example .env
# Edit .env with your settings
```

### 3. Start with Docker (Recommended)
```bash
# Start everything
make docker-up

# Or just infra (for local Go dev)
make infra-up
make run
```

### 4. Manual Run
```bash
go mod download
make run
# Server starts at http://localhost:8080
```

---

## 📡 API Reference

Base URL: `http://localhost:8080/api/v1`

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | Login with email/mobile + password |
| POST | `/auth/register` | Register new user |
| POST | `/auth/google` | Google OAuth login |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Logout (revoke refresh token) |
| GET  | `/auth/me` | Get current user |
| PUT  | `/auth/change-password` | Change password |

### Dashboard
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dashboard/stats?date=YYYY-MM-DD&outlet_id=` | Dashboard stats |
| GET | `/dashboard/outlet-stats?date=YYYY-MM-DD` | Per-outlet breakdown |
| GET | `/dashboard/orders-chart?date=&tab=orders|sales|net_sales|tax|discounts` | Chart data |
| GET | `/dashboard/summary?from=&to=` | Period summary |

### Outlets
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/outlets` | List user's outlets |
| POST | `/outlets` | Create outlet (admin/owner) |
| GET | `/outlets/:id` | Get outlet with zones |
| PUT | `/outlets/:id` | Update outlet |
| DELETE | `/outlets/:id` | Delete outlet |
| PATCH | `/outlets/:id/lock` | Toggle lock |
| GET | `/outlets/:id/zones` | Get zones & tables |
| POST | `/outlets/:id/zones` | Create zone |

### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/orders` | List orders (filters: outlet_id, status, source, from, to) |
| POST | `/orders` | Create order |
| GET | `/orders/running` | Live/running orders |
| GET | `/orders/online` | Online orders (Zomato, Swiggy, etc.) |
| GET | `/orders/:id` | Order details |
| PATCH | `/orders/:id/status` | Update status |
| PATCH | `/orders/:id/cancel` | Cancel order |
| POST | `/orders/:id/print` | Mark printed / reprint |

### Menu
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/menu/categories?outlet_id=` | List categories |
| POST | `/menu/categories` | Create category |
| GET | `/menu/items?outlet_id=&category_id=&is_available=` | List items |
| POST | `/menu/items` | Create item |
| GET | `/menu/out-of-stock?outlet_id=` | Out-of-stock items |
| PATCH | `/menu/items/:id/availability` | Toggle availability |
| PATCH | `/menu/items/:id/online` | Toggle online status |

### Reports
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/reports/list` | Report catalogue |
| GET | `/reports/sales?outlet_ids=&from=&to=&status=` | Sales summary report |

### Purchases
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/purchases/pending` | Pending purchase orders |
| POST | `/purchases` | Create purchase |
| PATCH | `/purchases/:id/status` | Update status |

### Notifications
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/notifications` | List notifications |
| PATCH | `/notifications/:id/read` | Mark one read |
| PATCH | `/notifications/read-all` | Mark all read |

### Third-Party Config (Admin/Owner)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/thirdparty?outlet_id=` | List configs |
| PUT | `/thirdparty/:id` | Update config (API key, store ID) |

### Logs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/logs/menu-triggers` | Menu trigger logs |
| GET | `/logs/online-store` | Online store toggle logs |
| GET | `/logs/online-items` | Item online status logs |

### Franchises (Admin/Owner)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/franchises` | List franchises |
| POST | `/franchises` | Create franchise |

### Users (Admin)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/billers` | List billers |
| GET | `/users/admins` | List admins |
| POST | `/users/invite` | Invite user |
| PUT | `/users/:id` | Update user |
| DELETE | `/users/:id` | Delete user |

---

## 🔑 Auth

All protected routes require:
```
Authorization: Bearer <access_token>
```

**Roles:** `admin` > `owner` > `biller`

---

## 🗄 Database Schema

Key tables: `users`, `outlets`, `zones`, `tables`, `categories`, `menu_items`, `orders`, `order_items`, `payments`, `pending_purchases`, `menu_trigger_logs`, `online_store_logs`, `online_item_logs`, `third_party_configs`, `notifications`, `franchises`, `outlet_accesses`, `refresh_tokens`

Auto-migrated on startup via GORM.

---

## 🧪 Example Request

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email_or_mobile":"admin@example.com","password":"password123"}'
```

**Dashboard stats:**
```bash
curl http://localhost:8080/api/v1/dashboard/stats?date=2026-03-11 \
  -H "Authorization: Bearer <token>"
```

**Create order:**
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "outlet_id": "uuid",
    "type": "dine_in",
    "pax": 2,
    "items": [{"menu_item_id": "uuid", "quantity": 2}],
    "payments": [{"method": "cash", "amount": 500}]
  }'
```

---

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22 |
| Framework | Gin v1.10 |
| ORM | GORM v2 + PostgreSQL driver |
| Auth | JWT (golang-jwt/jwt v5) |
| Cache / Sessions | Redis (go-redis v9) |
| Password | bcrypt |
| IDs | UUID v4 (google/uuid) |
| Config | godotenv |
| Container | Docker + Docker Compose |
