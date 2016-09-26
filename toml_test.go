package toml

import (
	"github.com/pelletier/go-toml"
	"reflect"
	"testing"
	"fmt"
	"log"
)

//type Config struct {
//	User          string
//	Password      string
//	MaxConnection int64
//	Timeout       float64
//	ShowSlowQuery bool
//	//Addresses     []string
//	Postgres      Postgres `toml:postgres`
//}
//
//type Postgres struct {
//	User     string `toml:user`
//	Password string `toml:password`
//}
//
//func TestLoad(t *testing.T) {
//	c := &Config{}
//	e := Load(c, "test/config.toml", "development")
//	if e != nil {
//		t.Error(e)
//	}
//	log.Println(c)
//}

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
		fmt.Errorf("failed to load user data: %v", v)
	}

	// production
	v, e = getValue(stringType, tree, "user", "production")
	if e != nil {
		t.Fatal(e)
	}
	if v.Interface().(string) != "admin" {
		fmt.Errorf("failed to load user data: %v", v)
	}
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
	vi, ok := v.Interface().(Conf)
	if !ok || vi.User != "root" {
		fmt.Errorf("failed to load user data: %v", v)
	}
	log.Println(v)

	// production
	v, e = getValue(structType, tree, "database", "production")
	if e != nil {
		t.Fatal(e)
	}
	vi, ok = v.Interface().(Conf)
	if !ok || vi.User != "root" {
		fmt.Errorf("failed to load user data: %v", v)
	}
	log.Println(v)
}

/*
func TestLoadData_struct(t *testing.T) {
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

	type Database struct {
		User string
		Pass string
	}

	type Conf struct {
		Database Database
	}

	c := &Conf{}

	loadData(getFieldValue(c, "Database", "User"), tree, "database", "user", "development")
	if c.Database.User != "root" {
		log.Println(c)
		t.Error("failed to load user data")
	}

	loadData(getFieldValue(c, "Database", "User"), tree, "database", "user", "production")
	if c.Database.User != "admin" {
		log.Println(c)
		t.Error("failed to load user data")
	}
}

func TestLoadData_array(t *testing.T) {
	tree, e := toml.Load(`
	users = ["admin", "user1"]
	[development]
	users = ["root"]
	`)
	if e != nil {
		t.Fatal(e)
	}

	type Conf struct {
		Users []string
	}

	c := &Conf{}

	loadData(getFieldValue(c, "Users"), tree, "", "users", "development")
	if len(c.Users) != 1 || c.Users[0] != "root" {
		log.Println(c)
		t.Error("failed to load user data")
	}

	loadData(getFieldValue(c, "Users"), tree, "", "users", "production")
	if len(c.Users) != 2 || c.Users[0] != "admin" {
		log.Println(c)
		t.Error("failed to load user data")
	}
}

func TestLoadData_arraystruct(t *testing.T) {
	tree, e := toml.Load(`
	[[databases]]
	user = "user1"
	pass = "pass1"

	[[databases]]
	user = "user2"
	pass = "pass2"

	[[databases.development]]
	user = "userdev"
	pass = "passdev"
	`)
	if e != nil {
		t.Fatal(e)
	}

	type Database struct {
		User string
		Pass string
	}

	type Conf struct {
		Databases []Database
	}

	c := &Conf{}

	loadData(getFieldValue(c, "Databases"), tree, "", "databases", "development")
	if len(c.Databases) != 1 || c.Databases[0].User != "userdev" {
		log.Println("arraystruct:", c)
		t.Error("failed to load user data")
	}

	loadData(getFieldValue(c, "Databases"), tree, "", "databases", "production")
	if len(c.Databases) != 2 || c.Databases[0].User != "user1" {
		log.Println("arraystruct:"+
			"", c)
		t.Error("failed to load user data")
	}
}
*/
