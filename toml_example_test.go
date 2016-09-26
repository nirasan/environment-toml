package toml_test

import (
	"fmt"
	"github.com/nirasan/environment-toml"
	"log"
	"time"
)

type Example1Config struct {
	Int1    int64
	Float1  float64
	String1 string
	Bool1   bool
	Date1   time.Time
	Array1  []int64
}

/*
# test/example1.toml
int1 = 1
float1 = 0.1
string1 = "string 1"
bool1 = true
date1 = 1980-01-01T00:00:00Z
array1 = [1, 2, 3]

[development]
int1 = 2

[production]
float1 = 0.5
array1 = [4, 5]
*/

func ExampleLoad_example1development() {
	c := &Example1Config{}


	err := toml.Load(c, "test/example1.toml", "development")
	if err != nil {
		log.Fatal(err)
	}

	// `Int1` is overwritten
	fmt.Printf("%+v", c)

	// Output:
	// &{Int1:2 Float1:0.1 String1:string 1 Bool1:true Date1:1980-01-01 00:00:00 +0000 UTC Array1:[1 2 3]}
}

func ExampleLoad_example1production() {
	c := &Example1Config{}

	err := toml.Load(c, "test/example1.toml", "production")
	if err != nil {
		log.Fatal(err)
	}

	// `Float1` and `Array1` is overwritten
	fmt.Printf("%+v", c)

	// Output:
	// &{Int1:1 Float1:0.5 String1:string 1 Bool1:true Date1:1980-01-01 00:00:00 +0000 UTC Array1:[4 5]}
}



