# Programming

## Concorrency Control
In scenarios with concurrent updates in a database, the concept is called **optimistic concurrency control**, where you use a versioning mechanism (often a column like version_id) to ensure that changes are only applied if no other updates have occurred since the data was retrieved.

**PostgreSQL**, along with **pgxpool**, does support native features for handling concurrency, but adding a version column can still be useful in certain cases.

When handling concurrent updates in PostgreSQL using `pgxpool`, here are three options to ensure data consistency:

#### 1. **Optimistic Concurrency Control (Version Column)**

Use a `version_id` column that increments with each update, ensuring no concurrent modifications conflict.

**SQL Example:**

```sql
UPDATE clients
SET balance = $1, version_id = version_id + 1
WHERE id = $2 AND version_id = $3;
```

**Go Example:**
```go
func (r *ClientRepository) UpdateBalance(ctx context.Context, clientID int, newBalance, versionID int) error {
    query := `UPDATE clients SET balance = $1, version_id = version_id + 1 WHERE id = $2 AND version_id = $3`
    result, err := r.db.Exec(ctx, query, newBalance, clientID, versionID)
    if result.RowsAffected() == 0 {
        return errors.New("concurrent update detected")
    }
    return err
}
```

---

#### 2. **Row-Level Locks (SELECT FOR UPDATE)**

Lock the rows when reading to prevent other transactions from modifying them until your transaction is done.

**SQL Example:**
```sql
SELECT balance FROM clients WHERE id = $1 FOR UPDATE;
```

**Go Example:**
```go
func (r *ClientRepository) UpdateBalanceWithLock(ctx context.Context, clientID, newBalance int) error {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // Rollback if commit is not called

    row := tx.QueryRow(ctx, `SELECT balance FROM clients WHERE id = $1 FOR UPDATE`, clientID)
    var balance int
    if err := row.Scan(&balance); err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `UPDATE clients SET balance = $1 WHERE id = $2`, newBalance, clientID)
    return tx.Commit(ctx)
}
```

---

#### 3. **Serializable Transactions**

Use the serializable isolation level to enforce strict transaction ordering, ensuring no concurrent conflicts.

**Go Example:**
```go
func (r *ClientRepository) UpdateBalanceSerializable(ctx context.Context, clientID, newBalance int) error {
    tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Perform your read and write operations here
    row := tx.QueryRow(ctx, `SELECT balance FROM clients WHERE id = $1`, clientID)
    var balance int
    err = row.Scan(&balance)
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `UPDATE clients SET balance = $1 WHERE id = $2`, newBalance, clientID)
    return tx.Commit(ctx)
}
```

These options help handle concurrency issues when multiple requests update the same data, ensuring data integrity in PostgreSQL.

## Framework and Context
Go differentiates between concurrent requests by leveraging **goroutines** and the way **`context.Context`** is designed to propagate cancellation signals, deadlines, and other request-specific values. Here's how it works in detail:

### Key Concepts:

1. **Goroutines for Each Request:**
   When using a web framework like Gin (or the `net/http` package), each incoming HTTP request is handled in a **separate goroutine**. This means that every request runs independently, and even if two requests are handled simultaneously, they won't interfere with each other unless explicitly synchronized (e.g., with shared variables or locks).

2. **Request-Specific `context.Context`:**
   The `context.Context` that comes from `c.Request.Context()` is unique to each HTTP request. Gin (and the underlying `net/http` package) creates a new `context.Context` for each incoming request, which is linked to the lifecycle of that request.

   - The `context.Context` carries data, deadlines, and cancellation signals specific to the request that it originated from. So, when you call `c.Request.Context()`, you're getting a `context` that is tied to the current HTTP request only.
   
   - If the request is cancelled (e.g., if the client disconnects), the associated context is automatically cancelled, and any operations relying on that context (such as database queries or other IO operations) will be notified.

### How Go Keeps Requests Separate:

- **Goroutines Are Isolated:** Each request is handled in a separate goroutine. This ensures that the state, variables, and contexts for one request are completely isolated from others.
  
- **Context Tree:** Every `context.Context` can have a parent-child relationship. The `c.Request.Context()` is a child of a root context created for each HTTP request, and this chain of contexts is unique to the request. Go tracks this structure, so any values or cancellations in one request’s context won’t affect another request.

### Example Flow:

1. A user makes a request to your Gin API.
2. Gin spawns a new goroutine to handle the request, and within this goroutine, it creates a **request-specific `context.Context`**.
3. You call `ctx := c.Request.Context()`, which refers to the context tied to this specific request.
4. If you pass this context to your service layer or database layer, all operations are tied to this specific request.
5. If the client disconnects or a timeout occurs, the request's context is **automatically cancelled**, and any ongoing operations tied to that context will be stopped.

### Visualization:
- Request 1 -> Goroutine A -> `context.Context A`
- Request 2 -> Goroutine B -> `context.Context B`

Each context is tied to its respective request, and they are handled independently by their respective goroutines.

### Context Example:
If a client cancels their request, the context gets cancelled automatically:

```go
func (h *ClientHandler) CreateTransaction(c *gin.Context) {
    ctx := c.Request.Context()

    err := h.service.LongRunningOperation(ctx, data)
    
    if errors.Is(err, context.Canceled) {
        c.JSON(http.StatusRequestTimeout, gin.H{"error": "request cancelled"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transaction created successfully"})
}
```

In this example, if the client disconnects or the request times out, the context will be cancelled, and the service layer will know to stop processing. This happens independently for each request due to Go’s isolation through goroutines and request-specific contexts.

## Error Scope with Transactions
I was in a bug in this code where the rollback was never executed:
```go
	var err error

	defer func() {
        // 3. Know here, then exit because err was EQUALS nil.
		if err != nil {
			s.logger.Error("rolling back transaction because of error", "error", err)
			err := s.repo.Rollback(ctx)
			if err != nil {
				s.logger.Error("failed to rollback transaction", "error", err)
			}
		}
	}()

    ...

    // 1. This would raise an error
	if err:= s.repo.UpdateClientBalance(ctx, t.ClientID, t.Amount); err != nil {
        // 2. The runtime would enter here
		s.logger.Error("failed to update client balance", "error", err)
		return nil, err
	}
```

This happened because `err` is scoped to the `if` statement. I can scope it in the function level by doing:
```go
	var err error

    ...

    err = s.repo.UpdateClientBalance(ctx, t.ClientID, t.Amount)
	if err != nil {
		s.logger.Error("failed to update client balance", "error", err)
		return nil, err
	}
```

Know the `err` has the error value and the defer function can perform the rollback.

## Middlewares in API
I didn't had a use case for middlewares until I wanted to timeout a context for every request, it was nice to know how it works.