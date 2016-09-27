# environment-toml

Go library for environment specific settings loader defined by [TOML](https://github.com/mojombo/toml).

# Features

* Load TOML and return setting struct.
* When environment specific setting defined, overwrite default settings.

# Usage

## Import

```go
import "github.com/nirasan/environment-toml"
```

## Define setting file

```toml:config.toml
[user]
name = "user1"
age = 10

[user.development]
name = "user2"

[user.production]
name = "user3"
age = 20
```

## Define setting struct

```go
type Config struct {
    User User
}

type User struct {
    Name string
    Age  int64
}
```

## Load development setting

```go
// `development` environment
c := &Config{}
err := toml.Load(c, "config.toml", "development")
if err != nil {
    log.Fatal(err)
}
fmt.Println( c.User.Name ) //=> user2 | user.name overwritten by user.development.name
fmt.Println( c.User.Age ) //=> 10 | user.age
```

## Load production setting

```go
// `production` environment
c := &Config{}
err := toml.Load(c, "config.toml", "production")
if err != nil {
    log.Fatal(err)
}
fmt.Println( c.User.Name ) //=> user3 | user.name overwritten by user.production.name
fmt.Println( c.User.Age ) //=> 20 | user.age overwritten by user.production.age
```

# Examples

[Basic types (environment: development)](https://godoc.org/github.com/nirasan/environment-toml#example-Load--Example1development)

[Basic types (environment: production)](https://godoc.org/github.com/nirasan/environment-toml#example-Load--Example1production)

[Struct type (environment: development)](https://godoc.org/github.com/nirasan/environment-toml#example-Load--Example2development)

[Struct type (environment: production)](https://godoc.org/github.com/nirasan/environment-toml#example-Load--Example2production)

# License

MIT
