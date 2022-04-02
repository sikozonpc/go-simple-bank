## Simple bank


### Local Setup
```bash
# Start PostgreSQL docker image
make db-start
# Init the project DB
make init-db
# Run migrations
make migrate-up

# Check if tests are passing
make test
```

### Structure

The `db/query` SQL files are the source of truth of the `db/sqlc` folder, where these files are auto-generated from `make sqlc` after updating, it basically generates GO code from SQL.