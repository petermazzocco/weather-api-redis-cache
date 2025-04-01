# weather-api-redis-cache

Roadmap.sh url - https://roadmap.sh/projects/weather-api-wrapper-service

## How to run

1. Clone repository

```bash
git clone https://github.com/petermazzocco/weather-api-redis-cache
```

2. Get API key from Visual Crossing
3. Add the following env to your `.env.`

```
REDIS_ADDR=127.0.0.1:6379  ##local
REDIS_PASSWORD=
REDIS_DB=0
```

4. Run server

```go
go run .
```

5. Open browser at `localhost:8080`
