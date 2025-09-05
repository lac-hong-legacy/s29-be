# s29-be

## Backend Setup Guide


#### Start the database
If you have Docker installed, run:
```bash
docker-compose up -d
```

#### 3. Install dependencies
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

#### 4. Run database migrations
```goose -dir ./migrations status```
```goose -dir ./migrations up```
```goose -dir ./migrations down```
```goose -dir ./migrations reset```

### Environment Variables
- `GOOSE_DRIVER=postgres`
- `GOOSE_DBSTRING=postgres://postgres:postgres@localhost:5432/s29?sslmode=disable`

**Rebuild docker image**
- ```docker build -t s29-api .```
- ```
  docker run -d \
  --name s29-api \
  -p 8080:8080 \
  -v $(pwd):/app \
  --network s29-network \ 
  -e APP_ENV=development \
  -e APP_PORT=8080 \
  -e AUDORA_DB_HOST=s29-db \
  -e AUDORA_DB_PORT=5432 \
  -e AUDORA_DB_USER=postgres \
  -e AUDORA_DB_PASSWORD=postgres \
  -e AUDORA_DB_NAME=s29 \
  -e KRATOS_PUBLIC_URL=http://kratos:4433 \
  -e KRATOS_ADMIN_URL=http://kratos:4434 \
  -e JWT_SECRET=your-jwt-secret-here \
  s29-api
  ```