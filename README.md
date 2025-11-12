# SQuery - Spanner Query Executor

`squery` is a lightweight, generic wrapper around the official Google Cloud Spanner client for Go (`cloud.google.com/go/spanner`). It simplifies common query execution patterns, reduces boilerplate, and provides a type-safe way to map query results to Go structs.

## Features

*   **Generic Executor**: A type-safe, generic `Executor[E]` to map rows to your specific Go struct type `E`.
*   **Query Builder Integration**: Works seamlessly with [Squirrel](https://github.com/Masterminds/squirrel) and includes helper functions to get started quickly.
*   **Simplified Row & Rows Fetching**: Easy-to-use methods like `Row()`, `Rows()`, and `Col()` for fetching single rows, multiple rows, or single column values.
*   **Automatic Parameter Naming**: Automatically converts `?` style SQL arguments into Spanner's named parameters (`@p1`, `@p2`, etc.).
*   **Transaction Support**: Execute queries within an existing read-write transaction.
*   **Standard Error Handling**: Returns `spanner.ErrRowNotFound` when a single row is expected but none is found.

## Installation

```bash
go get github.com/maximorov/squery
go get github.com/Masterminds/squirrel
```

## Quick Start

### 1. Define Your Entity

Your data model struct must implement the `squery.Entity` interface. This is done by adding a `ToData()` method that returns pointers to the struct's fields in the order they appear in your `SELECT` statement.

```go
package main

import (
    "cloud.google.com/go/spanner"
    "github.com/maximorov/squery"
)

type User struct {
    ID   int64
    Name string
    Age  spanner.NullInt64
}

// ToData returns pointers to the struct fields for scanning.
func (u *User) ToData() []any {
    return []any{&u.ID, &u.Name, &u.Age}
}
```

### 2. Build Queries with Squirrel

`squery` is designed to work perfectly with `squirrel`. The `squirrel.SelectBuilder` already implements the `squery.SqSQLer` interface, so you can use it directly.

`squery` provides a helper function `squery.Select()` which is a drop-in replacement for `squirrel.Select()`. It automatically sets the placeholder format to `squirrel.AtP`, which is required for Google Spanner.

```go
import (
    "github.com/maximorov/squery"
    sq "github.com/Masterminds/squirrel"
)

// Create a query using the squery helper
query := squery.Select("UserID", "Name", "Age").
    From("Users").
    Where(sq.Gt{"Age": 30})

// The above is equivalent to:
query = sq.Select("UserID", "Name", "Age").
    PlaceholderFormat(sq.AtP).
    From("Users").
    Where(sq.Gt{"Age": 30})
```

### 3. Execute Queries

Create a Spanner client, instantiate the `squery.Executor`, build a query with Squirrel, and start fetching data.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "cloud.google.com/go/spanner"
    "github.com/maximorov/squery"
    sq "github.com/Masterminds/squirrel"
)

// ... (User struct from above)

func main() {
    ctx := context.Background()
    // Assume spannerClient is an initialized *spanner.Client
    var spannerClient *spanner.Client 

    // Create a new executor for the User type.
    userExecutor := squery.NewExecutor[User](spannerClient)

    // --- Fetch multiple rows ---
    query := squery.Select("UserID", "Name", "Age").
        From("Users").
        Where(sq.Gt{"Age": 30})
        
    users, err := userExecutor.Rows(ctx, query)
    if err != nil {
        log.Fatalf("Failed to get users: %v", err)
    }
    fmt.Printf("Found %d users over 30.\n", len(users))

    // --- Fetch a single row ---
    query = squery.Select("UserID", "Name", "Age").
        From("Users").
        Where(sq.Eq{"UserID": 1})

    user, err := userExecutor.Row(ctx, query)
    if err != nil {
        if err == spanner.ErrRowNotFound {
            log.Println("User not found")
        } else {
            log.Fatalf("Failed to get user: %v", err)
        }
    }
    if user != nil {
        fmt.Printf("Found user: %s\n", user.Name)
    }

    // --- Fetch a single column from a single row ---
    // Note the generic type is now `string`
    nameExecutor := squery.NewExecutor[string](spannerClient)
    
    query = squery.Select("Name").
        From("Users").
        Where(sq.Eq{"UserID": 1})

    name, err := nameExecutor.Col(ctx, query)
    if err != nil {
        log.Fatalf("Failed to get user name: %v", err)
    }
    fmt.Printf("The name of user 1 is: %s\n", name)
}
```
