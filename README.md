## Simple bank

### Local Setup
```bash
# Start PostgreSQL docker image
make db-start
# Init the project DB
make init-db
# Run migrations
make migrate-up
```