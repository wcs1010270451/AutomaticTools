# AutomaticTools Backend

Go backend for a paid utility-tool platform.

The first tool is `auto_click`, but the backend is structured so more paid tools can be added later.

## Current Capabilities

- User registration and login
- JWT authentication
- Tool catalog
- Order creation
- Permanent tool entitlements after payment confirmation
- Device binding
- Audit logs
- JSON request logs with request IDs
- PostgreSQL persistence

## Architecture

Core tables:

- `users`: account identity
- `tools`: paid utility tools, for example `auto_click`
- `orders`: purchase records
- `entitlements`: which user owns which tool
- `devices`: user device binding and last-seen tracking
- `audit_logs`: traceable business events

Recommended client flow:

1. User registers or logs in.
2. Client calls `GET /api/tools`.
3. Client creates an order for a tool with `POST /api/orders`.
4. Payment provider confirms payment.
5. Backend marks order paid and grants entitlement.
6. Client calls `GET /api/me/entitlements`.
7. Client unlocks owned tools such as `auto_click`.

The temporary admin confirm endpoint exists only until a real payment provider is integrated.

## Start

```powershell
cd D:\wcs\Code\AutomaticTools\backend
go run .\cmd\api
```

Or double click:

```text
D:\wcs\Code\AutomaticTools\backend\run_backend.bat
```

Default address:

```text
http://0.0.0.0:8088
```

Runtime configuration is stored in `config.json` in this directory. Start by
copying values from `config.example.json` and change the database password and
secrets. The default database connection is:

```text
postgres://postgres@localhost:5432/automatic_tools?sslmode=disable
```

Create the `automatic_tools` database, then execute the initialization script:

```text
D:\wcs\Code\AutomaticTools\backend\sql\001_init_postgres.sql
```

The backend also executes this idempotent script during startup. See
`sql/README.md` for the manual commands and field conventions.

## Configuration

`config.json` supports:

- `addr`
- `database_url`
- `jwt_secret`
- `admin_key`
- `token_ttl_hours`
- `log_level`

Environment variables override the file when needed. Set `AUTOMATIC_TOOLS_CONFIG_FILE`
to load a different JSON file.

```powershell
$env:AUTOMATIC_TOOLS_PORT="8088"
$env:DATABASE_URL="postgres://postgres:your-password@localhost:5432/automatic_tools?sslmode=disable"
$env:AUTOMATIC_TOOLS_JWT_SECRET="change-this-secret"
$env:AUTOMATIC_TOOLS_ADMIN_KEY="change-this-admin-key"
go run .\cmd\api
```

Change `AUTOMATIC_TOOLS_JWT_SECRET` and `AUTOMATIC_TOOLS_ADMIN_KEY` before any real deployment.

## API

### Health

```http
GET /health
```

### Register

```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "test001",
  "password": "123456",
  "email": "test001@example.com",
  "phone": "+8613800138000",
  "deviceId": "android-device-id",
  "deviceName": "Honor",
  "platform": "android"
}
```

`email` and `phone` are optional during registration. When provided, both must
be unique; email matching is case-insensitive. Login continues to use username
and password until email/SMS verification is implemented.

### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "test001",
  "password": "123456",
  "deviceId": "android-device-id",
  "deviceName": "Honor",
  "platform": "android"
}
```

### Current User

```http
GET /api/me
Authorization: Bearer <token>
```

### Tool Catalog

```http
GET /api/tools
```

Example response:

```json
{
  "tools": [
    {
      "code": "auto_click",
      "name": "自动点击",
      "description": "基础自动点击工具，购买后永久使用。",
      "priceCents": 1000,
      "currency": "CNY",
      "lifetime": true
    }
  ]
}
```

### Create Order

```http
POST /api/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "toolCode": "auto_click",
  "payChannel": "manual"
}
```

The order starts as `pending`.

### My Orders

```http
GET /api/me/orders
Authorization: Bearer <token>
```

### My Entitlements

```http
GET /api/me/entitlements
Authorization: Bearer <token>
```

The client should unlock a tool only when the matching `toolCode` is present.

### Bind Device

```http
POST /api/devices/bind
Authorization: Bearer <token>
Content-Type: application/json

{
  "deviceId": "android-device-id",
  "deviceName": "Honor",
  "platform": "android"
}
```

### Confirm Order (temporary admin API)

This simulates payment completion before WeChat/Alipay integration exists.

```http
POST /api/admin/orders/confirm
X-Admin-Key: <AUTOMATIC_TOOLS_ADMIN_KEY>
Content-Type: application/json

{
  "orderNo": "ord_xxx"
}
```

This marks the order as `paid` and grants a permanent entitlement for the purchased tool.

### Grant Entitlement (temporary admin API)

Manual entitlement grant, useful for testing or customer support.

```http
POST /api/admin/entitlements/grant
X-Admin-Key: <AUTOMATIC_TOOLS_ADMIN_KEY>
Content-Type: application/json

{
  "userId": 1,
  "toolCode": "auto_click",
  "source": "admin",
  "orderNo": "manual-001"
}
```

## Android Device Access

When the backend runs on your PC, Android must call the PC LAN IP, for example:

```text
http://192.168.1.10:8088
```

Do not use `localhost` on Android. On Android, `localhost` means the phone itself.

## Production Notes

Before launch, add:

- Real payment orders and callbacks
- Payment callback signature verification
- Admin console
- Database backup
- Rate limiting
- Structured error monitoring
- Deployment health checks
- Refund and customer-service workflows
