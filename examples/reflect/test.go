package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"reflect"
)

func main() {
	dsh := Dish{Id: 1001}
	dsh.Db.Nnn = "hahah"
	// iterate through the attributes of a Data Model instance
	for name, mtype := range attributes(&dsh) {
		fmt.Printf("Name: %s, Type %s\n", name, mtype.Name())
	}
	v := reflect.ValueOf(&dsh)
	v = v.Elem()
	fmt.Printf("Name: Id, Value %d, Kind %s\n", v.FieldByName("Id").Int(), v.FieldByName("Da").Kind())

	dbv := v.FieldByName("Db").Interface()
	b, err := json.Marshal(dbv)
	if err != nil {
		fmt.Printf("json DB err %s \n", err.Error())
	}
	fmt.Printf("json DB %s, %v \n", string(b), dbv)
	fmt.Printf("Da Is Structs? %v \n", structs.IsStruct(dsh.Da))
	fmt.Printf("Db Is Structs? %v \n", structs.IsStruct(dsh.Db))
}

type DA int

type DB struct {
	test int
	Nnn  string
}

// Data Model
type Dish struct {
	Id     int
	Name   string
	Origin string
	Da     DA
	Db     DB
	Query  func()
}

// Example of how to use Go's reflection
// Print the attributes of a Data Model
func attributes(m interface{}) map[string]reflect.Type {
	typ := reflect.TypeOf(m)
	// if a pointer to a struct is passed, get the type of the dereferenced object
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// create an attribute data structure as a map of types keyed by a string.
	attrs := make(map[string]reflect.Type)
	// Only structs are supported so return an empty result if the passed object
	// isn't a struct
	if typ.Kind() != reflect.Struct {
		fmt.Printf("%v type can't have attributes inspected\n", typ.Kind())
		return attrs
	}

	// loop through the struct's fields and set the map
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			attrs[p.Name] = p.Type
		}
	}

	return attrs
}
