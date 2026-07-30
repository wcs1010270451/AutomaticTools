# Database SQL

The backend uses PostgreSQL. `001_init_postgres.sql` is the canonical schema
executed automatically by `internal/store.Migrate` and can also be run manually.

## Initialize manually

Create the database first while connected to PostgreSQL's default `postgres`
database:

```powershell
psql -U postgres -d postgres -c "CREATE DATABASE automatic_tools WITH ENCODING 'UTF8';"
```

Then initialize its tables:

```powershell
psql -U postgres -d automatic_tools -f .\sql\001_init_postgres.sql
```

For a database created before the user email and phone fields were added, run:

```powershell
psql -U postgres -d automatic_tools -f .\sql\002_add_user_contacts_postgres.sql
```

On this computer, `psql.exe` is currently located at:

```text
D:\Users\wang1\Downloads\postgresql-17.5-3-windows-x64-binaries\pgsql\bin\psql.exe
```

If `psql` is not in `PATH`, call that executable with PowerShell's `&` operator,
or open the SQL file in pgAdmin's Query Tool.

The table script is idempotent: tables and indexes use `IF NOT EXISTS`, and the
default `auto_click` tool uses `ON CONFLICT DO NOTHING`.

## Field conventions

- Monetary amounts use integer cents, for example `1000` means CNY 10.00.
- Time fields use Unix timestamps in seconds.
- `tools.lifetime` and `tools.active` use PostgreSQL booleans.
- An entitlement with a null `expires_at` is permanent.
