# Programming

Here’s a separation between **optimistic concurrency control** and **pessimistic concurrency control**, along with clear explanations of each:

## Concurrency Control
In database systems where multiple users or processes may try to modify the same data concurrently, two common approaches are **optimistic concurrency control** and **pessimistic concurrency control**. Here's how each works:

### 1. **Optimistic Concurrency Control**
In optimistic concurrency control, the system assumes that conflicts between concurrent updates are rare. Therefore, it allows multiple transactions to proceed without any locks, and only checks for conflicts when trying to commit the transaction. A common technique is to use versioning (e.g., a `version_id` column) to detect if any concurrent updates have occurred.

#### **Use Case:**
- Suitable for systems where the likelihood of conflicts is low.
- It avoids locking the data, improving performance in low-contention environments.

#### Example with a Version Column:

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
In this example, the `version_id` ensures that updates are only applied if no other updates have occurred since the data was last read.

---

### 2. **Pessimistic Concurrency Control**
In pessimistic concurrency control, the system assumes that conflicts are likely, so it locks the data to prevent other transactions from making changes. This ensures that only one transaction can modify the data at a time, thus preventing conflicts at the cost of reduced concurrency.

#### **Use Case:**
- Suitable for high-contention environments where multiple processes are likely to attempt updates to the same data simultaneously.
- Ensures data integrity by using locks, but may result in slower performance due to blocking.

#### 2.1 **Row-Level Locks (SELECT FOR UPDATE)**

In this approach, you lock the rows when reading them using `SELECT ... FOR UPDATE`, preventing other transactions from modifying the locked rows until your transaction completes.

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
    defer tx.Rollback(ctx)

    row := tx.QueryRow(ctx, `SELECT balance FROM clients WHERE id = $1 FOR UPDATE`, clientID)
    var balance int
    if err := row.Scan(&balance); err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `UPDATE clients SET balance = $1 WHERE id = $2`, newBalance, clientID)
    return tx.Commit(ctx)
}
```
This locks the row during the transaction, preventing other transactions from modifying it until the lock is released.

---

#### 2.2 **Serializable Transactions**
Another form of pessimistic control is to use the `Serializable` isolation level, which makes transactions appear as if they were executed sequentially, one after the other. This is the strictest isolation level, ensuring no concurrent transaction conflicts.

**Go Example:**
```go
func (r *ClientRepository) UpdateBalanceSerializable(ctx context.Context, clientID, newBalance int) error {
    tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    row := tx.QueryRow(ctx, `SELECT balance FROM clients WHERE id = $1`, clientID)
    var balance int
    if err := row.Scan(&balance); err != nil {
        return err
    }

    _, err = tx.Exec(ctx, `UPDATE clients SET balance = $1 WHERE id = $2`, newBalance, clientID)
    return tx.Commit(ctx)
}
```
This ensures that your transaction will either run without any conflicts or will be retried if a conflict is detected.

---

### Conclusion:
- **Optimistic Concurrency Control** is useful in low-contention environments, where the cost of handling occasional conflicts is less than the overhead of locking.
- **Pessimistic Concurrency Control** is better in high-contention environments where conflicts are more likely, and preventing them with locks or serializable transactions ensures data consistency at the cost of concurrency.

Both approaches help handle concurrency issues, but the right choice depends on the specifics of your system and its contention level.

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

## Slices in Go
Slices are already reference types, meaning they act like pointers internally. When passing a slice, you’re not copying the underlying data—just the slice header (pointer, length, capacity).

Therefore:
```go
*[]domain.Transaction // This wouldn't make sense, but it works. I was using it.
```

But this could:
```go
[]*domain.Transaction // Nice option, but the structs are passed by value.
```

Finally, this is simplier and more readable:
```go
[]domain.Transaction // Generraly good. Pass the structs by value (wouldn't be great in large dataset).
```

Here’s the summary:

[]domain.Transaction (Slice of Structs):
- **Pros:** Simpler code, better memory locality (cache-friendly), no pointer dereferencing.
- **Cons:** Higher memory usage if structs are large, copies entire structs.

[]*domain.Transaction (Slice of Pointers):
- **Pros:** Lower memory usage (only copying pointers), avoids unnecessary struct copies, better for large structs.
- **Cons:** Requires pointer dereferencing, potential memory fragmentation, and more complex lifetime management.

Recommendation:
- Use []domain.Transaction if your structs are small and immutable.
- Use []*domain.Transaction if your structs are large or need to be modified frequently.