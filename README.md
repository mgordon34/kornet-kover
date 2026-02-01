# Kornet Kover

## Cloud-to-local sync (NBA)

Set the cloud Postgres URL and run the sync command.

```bash
export CLOUD_DB_URL="postgres://..."
export DB_URL="postgres://..."  # local database

go run ./backend/cmd/cloudsync
```
