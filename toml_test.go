package toml

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"log"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {

	type Postgres struct {
		User     string
		Password string
		Tables   []string
	}

	type Conf struct {
		User          string
		Password      string
		MaxConnection int64
		ShowSlowQuery bool
		Addresses     []string
		Postgres      Postgres
	}

	// development
	{
		c := &Conf{}
		err := Load(c, "test/config.toml", "development")
		if err != nil {
			t.Fatal(err)
		}
		if c.User != "master user" || c.Password != "12345" || c.MaxConnection != 1 || c.ShowSlowQuery != true || len(c.Addresses) != 2 || c.Addresses[0] != "192.168.0.1" {
			t.Error(fmt.Sprintf("failed to load conf: %v", c))
		}
		if c.Postgres.User != "root" || c.Postgres.Password != "" || len(c.Postgres.Tables) != 3 || c.Postgres.Tables[2] != "debuglog" {
			t.Error(fmt.Sprintf("failed to load struct: %v", c))
		}
		log.Println(c)
	}

	// production
	{
		c := &Conf{}
		err := Load(c, "test/config.toml", "production")
		if err != nil {
			t.Fatal(err)
		}
		if c.User != "master user" || c.Password != "master password" || c.MaxConnection != 100 || c.ShowSlowQuery != false || len(c.Addresses) != 3 || c.Addresses[0] != "10.0.0.1" {
			t.Error(fmt.Sprintf("failed to load conf: %v", c))
		}
		if c.Postgres.User != "rouser" || c.Postgres.Password != "mypassword" || len(c.Postgres.Tables) != 2 || c.Postgres.Tables[1] != "password" {
			t.Error(fmt.Sprintf("failed to load struct: %v", c))
		}
		log.Println(c)
	}
}

func TestGetValue_basic(t *testing.T) {
	tree, e := toml.Load(`
	user = "admin"
	pass = "admin"
	[development]
	user = "root"
	`)
	if e != nil {
		t.Fatal(e)
	}

	stringType := reflect.TypeOf("")

	// development
	v, e := getValue(stringType, tree, "user", "development")
	if e != nil {
		t.Fatal(e)
	}
	if v.Interface().(string) != "root" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}

	// production
	v, e = getValue(stringType, tree, "user", "production")
	if e != nil {
		t.Fatal(e)
	}
	if v.Interface().(string) != "admin" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
}

func TestGetValue_map(t *testing.T) {
	tree, e := toml.Load(`
	[database]
	user = "admin"
	pass = "admin"
	[database.development]
	user = "root"
	`)
	if e != nil {
		t.Fatal(e)
	}

	mapType := reflect.TypeOf(map[string]string{})

	// development
	v, e := getValue(mapType, tree, "database", "development")
	if e != nil {
		t.Fatal(e)
	}
	m, ok := v.Interface().(map[string]string)
	if !ok || m["user"] != "root" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)

	// production
	v, e = getValue(mapType, tree, "database", "production")
	if e != nil {
		t.Fatal(e)
	}
	m, ok = v.Interface().(map[string]string)
	if !ok || m["user"] != "admin" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)
}

func TestGetValue_basicarray(t *testing.T) {
	tree, e := toml.Load(`
	users = ["admin", "root"]
	[development]
	users = ["devuser"]
	`)
	if e != nil {
		t.Fatal(e)
	}

	arrayType := reflect.TypeOf([]string{})

	// development
	v, e := getValue(arrayType, tree, "users", "development")
	if e != nil {
		t.Fatal(e)
	}
	m, ok := v.Interface().([]string)
	if !ok || len(m) != 1 || m[0] != "devuser" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)

	// production
	v, e = getValue(arrayType, tree, "users", "production")
	if e != nil {
		t.Fatal(e)
	}
	m, ok = v.Interface().([]string)
	if !ok || len(m) != 2 || m[0] != "admin" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)
}

func TestGetValue_struct(t *testing.T) {
	tree, e := toml.Load(`
	[database]
	user = "admin"
	pass = "admin"
	[database.development]
	user = "root"
	`)
	if e != nil {
		t.Fatal(e)
	}

	type Conf struct {
		User string
		Pass string
	}

	structType := reflect.TypeOf(Conf{})

	// development
	v, e := getValue(structType, tree, "database", "development")
	if e != nil {
		t.Fatal(e)
	}
	s, ok := v.Interface().(Conf)
	if !ok || s.User != "root" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)

	// production
	v, e = getValue(structType, tree, "database", "production")
	if e != nil {
		t.Fatal(e)
	}
	s, ok = v.Interface().(Conf)
	if !ok || s.User != "admin" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println(v)
}

func TestGetValue_arraystruct(t *testing.T) {
	tree, e := toml.Load(`
	[[database]]
	user = "user1"
	pass = "pass1"
	[[database]]
	user = "user2"
	pass = "pass2"
	[[development.database]]
	user = "devuser1"
	pass = "devpass1"
	`)
	if e != nil {
		t.Fatal(e)
	}

	type Conf struct {
		User string
		Pass string
	}

	arrayStructType := reflect.TypeOf([]Conf{})

	// development
	v, e := getValue(arrayStructType, tree, "database", "development")
	if e != nil {
		t.Fatal(e)
	}
	s, ok := v.Interface().([]Conf)
	if !ok || len(s) != 1 || s[0].User != "devuser1" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println("array struct 1:", s)

	// production
	v, e = getValue(arrayStructType, tree, "database", "production")
	if e != nil {
		t.Fatal(e)
	}
	s, ok = v.Interface().([]Conf)
	if !ok || len(s) != 2 || s[0].User != "user1" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println("arrya struct 2:", s)
}

func TestGetValue_arraymap(t *testing.T) {
	tree, e := toml.Load(`
	[[database]]
	user = "user1"
	pass = "pass1"
	[[database]]
	user = "user2"
	pass = "pass2"
	[[development.database]]
	user = "devuser1"
	pass = "devpass1"
	`)
	if e != nil {
		t.Fatal(e)
	}

	arrayMapType := reflect.TypeOf([]map[string]string{})

	// development
	v, e := getValue(arrayMapType, tree, "database", "development")
	if e != nil {
		t.Fatal(e)
	}
	s, ok := v.Interface().([]map[string]string)
	if !ok || len(s) != 1 || s[0]["user"] != "devuser1" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println("array map 1:", s)

	// production
	v, e = getValue(arrayMapType, tree, "database", "production")
	if e != nil {
		t.Fatal(e)
	}
	s, ok = v.Interface().([]map[string]string)
	if !ok || len(s) != 2 || s[0]["user"] != "user1" {
		t.Error(fmt.Sprintf("failed to load user data: %v", v))
	}
	log.Println("arrya map 2:", s)
}
