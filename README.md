This backend is built with Golang (Fiber) and uses MongoDB to store notes.
It implements a fixed window counter rate limiter using Redis, limiting each IP to a fixed number of requests per minute.
Redis increments a counter per IP and sets a TTL so the count resets automatically after the window ends.
If the request count exceeds the limit within the time window, the API returns HTTP 429 Too Many Requests.

Steps to run : Frontend - python3 -m http.server 5500, 
Backend - go run main.go
