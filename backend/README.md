# AutomaticTools Backend

Gin + GORM backend for a paid utility-tool platform.

The first tool is `auto_click`, but the backend is structured so more paid tools can be added later.

## Current Capabilities

- User registration and login
- Administrator login with isolated JWT authorization
- Administrator, user, and tool management APIs
- JWT authentication
- Tool catalog
- Order creation
- Permanent tool entitlements after payment confirmation
- Device binding
- Audit logs
- JSON request logs with request IDs
- PostgreSQL persistence through GORM

## Architecture

Code layout:

- `internal/router`: Gin engine setup and route registration
- `internal/handler`: HTTP request parsing, authentication headers, and JSON responses
- `internal/logic`: framework-independent authentication, DTOs, and business logic
- `internal/middleware`: request IDs, CORS, access logging, and panic recovery
- `internal/store`: GORM models, PostgreSQL connection, and schema initialization

Core tables:

- `users`: account identity
- `admins`: management-console accounts
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
- `admin_username`
- `admin_password`
- `token_ttl_hours`
- `smtp_host`
- `smtp_port`
- `smtp_username`
- `smtp_password`
- `smtp_from`
- `smtp_from_name`
- `smtp_encryption` (`starttls`, `tls`, or `none`)
- `log_level`

Environment variables override the file when needed. Set `AUTOMATIC_TOOLS_CONFIG_FILE`
to load a different JSON file.

```powershell
$env:AUTOMATIC_TOOLS_PORT="8088"
$env:DATABASE_URL="postgres://postgres:your-password@localhost:5432/automatic_tools?sslmode=disable"
$env:AUTOMATIC_TOOLS_JWT_SECRET="change-this-secret"
$env:AUTOMATIC_TOOLS_ADMIN_USERNAME="admin"
$env:AUTOMATIC_TOOLS_ADMIN_PASSWORD="change-this-password"
$env:AUTOMATIC_TOOLS_SMTP_HOST="smtp.example.com"
$env:AUTOMATIC_TOOLS_SMTP_PORT="587"
$env:AUTOMATIC_TOOLS_SMTP_USERNAME="noreply@example.com"
$env:AUTOMATIC_TOOLS_SMTP_PASSWORD="smtp-authorization-code"
$env:AUTOMATIC_TOOLS_SMTP_FROM="noreply@example.com"
$env:AUTOMATIC_TOOLS_SMTP_ENCRYPTION="starttls"
go run .\cmd\api
```

On the first startup, the backend creates the administrator from these bootstrap
credentials. The defaults are `admin` / `123456`. Existing administrator
passwords are never overwritten during later startups. Change
`AUTOMATIC_TOOLS_JWT_SECRET` and `AUTOMATIC_TOOLS_ADMIN_PASSWORD` before any
real deployment.

## API

### Health

```http
GET /health
```

### Send Registration Email Code

```http
POST /api/auth/email-code
Content-Type: application/json

{
  "email": "test001@example.com"
}
```

The six-digit code is valid for 10 minutes. The same email can request another
code after 60 seconds. Only an HMAC-SHA256 hash is stored in PostgreSQL, and a
code is invalidated after five failed attempts or one successful registration.

SMTP must be configured before this endpoint can send mail. Use the mail
provider's SMTP authorization code when the provider does not permit the normal
mailbox password. Port 587 commonly uses `starttls`; port 465 commonly uses
implicit `tls`.

### Register

```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "test001@example.com",
  "emailCode": "123456",
  "password": "123456",
  "username": "test001",
  "phone": "+8613800138000",
  "deviceId": "android-device-id",
  "deviceName": "Honor",
  "platform": "android"
}
```

`email`, `emailCode`, and `password` are required. `username` is optional; the
backend generates an internal unique username when it is omitted. `phone` is
optional. Email and phone values must be unique, and email matching is
case-insensitive. Passwords must contain at least 6 characters and cannot
exceed bcrypt's 72-byte input limit.

### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "account": "test001",
  "password": "123456",
  "deviceId": "android-device-id",
  "deviceName": "Honor",
  "platform": "android"
}
```

`account` accepts a username, email address, or phone number. Existing clients
may continue sending `username` instead of `account`; it is retained as a
backward-compatible alias. A successful response contains the user JWT and the
current user record. Disabled users cannot log in, and their previously issued
tokens stop working immediately.

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

### Admin Login

```http
POST /api/admin/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456"
}
```

The returned token is an administrator token and cannot be used as an app-user
token. Likewise, app-user tokens cannot access management APIs.

### List Users

This API requires an administrator token:

```http
GET /api/admin/users?page=1&pageSize=20&search=test&status=active
Authorization: Bearer <admin-token>
```

`search` matches username, email, and phone. `status` can be `active` or
`disabled`; omit either filter to return all matching users. The response
contains `users`, `total`, `page`, and `pageSize`, and never exposes password
hashes.

### Manage Administrators

All administrator-management APIs require the administrator token:

```http
Authorization: Bearer <admin-token>
```

List administrators:

```http
GET /api/admin/admins
```

Create an administrator:

```http
POST /api/admin/admins
Content-Type: application/json

{
  "username": "admin_ops",
  "password": "123456",
  "status": "active"
}
```

Update an administrator. Leave `password` empty or omit it to keep the current
password:

```http
PUT /api/admin/admins/2
Content-Type: application/json

{
  "username": "admin_ops",
  "status": "disabled"
}
```

Delete an administrator:

```http
DELETE /api/admin/admins/2
```

The current administrator cannot delete or disable itself, and at least one
active administrator is always retained.

### Manage Tools

All tool-management APIs require an administrator token. List every tool,
including tools that are not currently visible to app users:

```http
GET /api/admin/tools
Authorization: Bearer <admin-token>
```

Create a tool:

```http
POST /api/admin/tools
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "code": "timer",
  "name": "计时器",
  "description": "简单的桌面计时工具。",
  "priceCents": 500,
  "currency": "CNY",
  "lifetime": true,
  "active": true
}
```

Update price, description, authorization type, or listing status:

```http
PUT /api/admin/tools/timer
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "name": "计时器",
  "description": "简单的桌面计时工具。",
  "priceCents": 800,
  "currency": "CNY",
  "lifetime": true,
  "active": false
}
```

Tool codes cannot be changed or deleted because orders and entitlements retain
their references to the code. Set `active` to `false` to remove a tool from the
public catalog. Create and update actions are written to `audit_logs`.

### Confirm Order (temporary admin API)

This simulates payment completion before WeChat/Alipay integration exists.

```http
POST /api/admin/orders/confirm
Authorization: Bearer <admin-token>
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
Authorization: Bearer <admin-token>
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
- Database backup
- Rate limiting
- Structured error monitoring
- Deployment health checks
- Refund and customer-service workflows
