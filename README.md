## QOR Admin

Instantly create a beautiful, cross platform, configurable Admin Interface and API for managing your data in minutes.

[![GoDoc](https://godoc.org/github.com/qor/admin?status.svg)](https://godoc.org/github.com/qor/admin)

**For security issues, please send us an email to security@getqor.com and give us time to respond BEFORE posting as an issue or reporting on public forums.**

## Features

- Generate Admin Interface for managing data
- RESTFul JSON API
- Association handling
- Search and filtering
- Actions/Batch Actions
- Authentication and Authorization
- Extendability

## Quick Start

```go
package main

import (
  "fmt"
  "net/http"
  "github.com/jinzhu/gorm"
  _ "github.com/mattn/go-sqlite3"
  "github.com/qor/admin"
)

// Create a GORM-backend model
type User struct {
  gorm.Model
  Name string
}

// Create another GORM-backend model
type Product struct {
  gorm.Model
  Name        string
  Description string
}

func main() {
  DB, _ := gorm.Open("sqlite3", "demo.db")
  DB.AutoMigrate(&User{}, &Product{})

  // Initialize
  Admin := admin.New(&admin.AdminConfig{DB: DB})

  // Allow to use Admin to manage User, Product
  Admin.AddResource(&User{})
  Admin.AddResource(&Product{})

  // initalize an HTTP request multiplexer
  mux := http.NewServeMux()

  // Mount admin interface to mux
  Admin.MountTo("/admin", mux)

  fmt.Println("Listening on: 9000")
  http.ListenAndServe(":9000", mux)
}
```

`go run main.go` and visit `localhost:9000/admin` to see the result!

## Live DEMO

* Live Demo [http://demo.getqor.com/admin](http://demo.getqor.com/admin)
* Source Code of Live Demo [https://github.com/qor/qor-example](https://github.com/qor/qor-example)

## Documentation

<https://doc.getqor.com/admin>

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
