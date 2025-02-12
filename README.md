# Rinha de Backend with Go (Edition 2024/Q1)

A high-performance REST API built with Go for the [Rinha de Backend 2024/Q1](https://github.com/zanfranceschi/rinha-de-backend-2024-q1) challenge - a competition focused on handling concurrent financial transactions with minimal resources (1.5 CPU cores and 550MB RAM).

## 🎯 Challenge Overview

Build a REST API that can handle:
- Processing credit/debit transactions for clients with high concurrency
- Retrieving account statements under heavy load
- Managing concurrent transactions with proper balance/limit validation
- Handling 3k+ simultaneous requests with limited resources

## 🔒 Concurrency Handling

This project demonstrates several concurrency patterns and safeguards:

1. **Database Concurrency**
   - Optimistic locking with `SELECT FOR UPDATE` to prevent race conditions
   - Connection pooling with `pgxpool` to manage concurrent database access
   - Transaction isolation to ensure data consistency

2. **API Concurrency**
   - Request timeouts via middleware to prevent resource exhaustion
   - Load balancing across multiple API instances
   - Goroutine management for concurrent request handling

3. **Resource Management**
   - Efficient memory usage with proper Go data structures
   - Connection pool limits to prevent database overload
   - Request throttling when under extreme load

## 🚀 Technical Stack

- **API Framework**: [Gin](https://github.com/gin-gonic/gin) for high-performance HTTP routing
- **Database**: PostgreSQL with `pgxpool` for efficient connection pooling
- **Load Balancer**: Nginx for request distribution
- **Architecture**: Clean Architecture principles
  - Domain-driven design
  - Repository pattern
  - Dependency injection
  - Middleware support

## 🏗️ Project Structure

```
├── cmd/
│   └── api/
│       ├── handler/    # HTTP request handlers
│       ├── middleware/ # Timeout and concurrency controls
│       ├── router/     # Route definitions
│       └── server.go   # Application entry point
├── internal/
│   ├── domain/        # Business logic and entities
│   └── infra/
│       ├── logger/    # Logging configuration
│       └── repository/# Concurrent data access layer
└── scripts/
    ├── nginx/         # Load balancer configuration
    └── postgres/      # Database initialization
```

## 🧪 Load Testing Results

The API successfully handles:
- 3k+ concurrent users
- Consistent response times under 100ms
- Zero data inconsistencies under load
- Full details in [LEARNING.md](./LEARNING.md)

## Getting Started
* Install Docker
* Run `docker compose up -d --build` or `make run`

## Stack
- **Gin** for REST API with concurrent request handling
- **Postgres** for Database with transaction isolation
- **pgxpool** driver for concurrent Postgres connections
- **Nginx** for load balancing across API instances

## References
- https://github.com/zanfranceschi/rinha-de-backend-2024-q1