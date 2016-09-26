package toml

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"go/ast"
	"reflect"
	"unicode"

	"errors"
)

func Load(v interface{}, file string, env string) error {
	tree, err := toml.LoadFile(file)
	if err != nil {
		return err
	}

	if v == nil {
		return fmt.Errorf("v must not be nil")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct pointer")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)
		if !ast.IsExported(ft.Name) {
			continue
		}
		name := getFieldName(ft)
		value, err := getValue(fv.Type(), tree, name, env)
		if err != nil {
			return err
		}
		fv.Set(value)
	}
	return nil
}

//func getData(t reflect.Type, tree *toml.TomlTree, path, elem, env string) reflect.Value {
//	envPath := createPath(path, env, elem)
//	elemPath := createPath(path, elem)
//	switch t.Kind() {
//	case reflect.Struct:
//		rv := reflect.New(t)
//		for i := 0; i < t.NumField(); i++ {
//			fv := rv.Field(i)
//			ft := t.Field(i)
//			if !ast.IsExported(ft.Name) {
//				continue
//			}
//			name := getFieldName(ft)
//			rv.Field(i).Set(getData(fv, tree, elemPath, name, env))
//		}
//		return rv
//	case reflect.Array, reflect.Slice:
//		rv := reflect.New(t).Elem()
//		p := elemPath
//		if tree.Has(envPath) {
//			p = envPath
//		}
//		node := tree.Get(p)
//		switch val := node.(type) {
//		// for primitive array
//		case []interface{}:
//			for _, el := range val {
//				rv = reflect.Append(rv, reflect.ValueOf(el))
//			}
//		// for map of array
//		case []*toml.TomlTree:
//			for i, childTree := range val {
//				value := getData(t.Elem(), childTree, "", , env)
//				log.Printf("%T %v", value, value)
//				rv = reflect.Append(rv, value)
//			}
//		default:
//		}
//		return rv
//	default:
//		if tree.Has(envPath) {
//			return reflect.ValueOf(tree.Get(envPath))
//		} else if tree.Has(elemPath) {
//			return reflect.ValueOf(tree.Get(elemPath))
//		}
//	}
//	return reflect.ValueOf(nil)
//}

var nilValue = reflect.ValueOf(nil)

func getValue(t reflect.Type, tree *toml.TomlTree, elem, env string) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Struct:
		return getStructValue(t, tree, elem, env)
	case reflect.Map:
		return getMapValue(t, tree, elem, env)
	case reflect.Array, reflect.Slice:
		return getArrayValue(t, tree, elem, env)
	default:
		return getBasicValue(t, tree, elem, env)
	}
}

func getBasicValue(t reflect.Type, tree *toml.TomlTree, elem, env string) (reflect.Value, error) {
	p, err := findPath(tree, elem, env)
	if err != nil {
		return nilValue, err
	}
	v := tree.Get(p)
	if reflect.TypeOf(v) != t {
		return nilValue, errors.New("invalid type")
	}
	return reflect.ValueOf(v), nil
}

func getArrayValue(t reflect.Type, tree *toml.TomlTree, elem, env string) (reflect.Value, error) {
	if t.Kind() != reflect.Array || t.Kind() != reflect.Slice {
		return nilValue, errors.New("invalid type")
	}
	p, err := findPath(tree, elem, env)
	if err != nil {
		return nilValue, err
	}
	v := tree.Get(p)
	et := t.Elem()
	rv := reflect.New(t).Elem()
	
	switch ary := v.(type) {
	case []*toml.TomlTree:
		for _, childTree := range ary {
			ev, e := getValue(et, childTree, "", elem)
			if e != nil {
				return nilValue, e
			}
			reflect.Append(rv, ev)
		}
	case []interface{}:
		for _, a := range ary {
			reflect.Append(rv, reflect.ValueOf(a))
		}
	default:
		return nilValue, errors.New("Invalid type")
	}
	return rv, nil
}

func getMapValue(t reflect.Type, tree *toml.TomlTree, elem, env string) (reflect.Value, error) {
	if t.Kind() != reflect.Map {
		return nilValue, errors.New("invalid type")
	}
	target := tree
	// find env if elem defined
	if elem != "" {
		p, err := findPath(tree, elem, env)
		if err != nil {
			return nilValue, err
		}
		newTree := tree.Get(p)
		switch val := newTree.(type) {
		case *toml.TomlTree:
			target = val
		default:
			return nilValue, errors.New("invalud tree")
		}
	}
	// get map value from tree
	rv := reflect.New(t).Elem()
	for _, k := range target.Keys() {
		kv := reflect.ValueOf(k)
		rv.SetMapIndex(kv, reflect.ValueOf(target.Get(k)))
	}
	return rv, nil
}

func getStructValue(t reflect.Type, tree *toml.TomlTree, elem, env string) (reflect.Value, error) {
	if t.Kind() != reflect.Struct {
		return nilValue, errors.New("invalid type")
	}
	target := tree
	// find env if elem defined
	if elem != "" {
		p, err := findPath(tree, elem, env)
		if err != nil {
			return nilValue, err
		}
		newTree := tree.Get(p)
		switch val := newTree.(type) {
		case *toml.TomlTree:
			target = val
		default:
			return nilValue, errors.New("invalid tree")
		}
	}
	// get struct value from tree
	rv := reflect.New(t).Elem()
	for i := 0; i < t.NumField(); i++ {
		fv := rv.Field(i)
		ft := t.Field(i)
		if !ast.IsExported(ft.Name) {
			continue
		}
		name := getFieldName(ft)
		value, err := getValue(ft.Type, target, name, env)
		if err != nil {
			return nilValue, err
		}
		fv.Set(value)
	}
	return rv, nil
}

func getFieldName(f reflect.StructField) string {
	name := f.Tag.Get("toml")
	if name == "" {
		name = toSnake(f.Name)
	}
	return name
}

func toSnake(in string) string {
	runes := []rune(in)
	l := len(runes)

	var out []rune
	for i := 0; i < l; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < l && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

func createPath(in ...string) string {
	out := ""
	for _, s := range in {
		if s != "" {
			if out == "" {
				out = s
			} else {
				out = out + "." + s
			}
		}
	}
	return out
}

func findPath(tree *toml.TomlTree, elem, env string) (string, error) {
	envPath := createPath(env, elem)
	elemPath := createPath(elem)
	if tree.Has(envPath) {
		return envPath, nil
	} else if tree.Has(elemPath) {
		return elemPath, nil
	} else {
		return "", errors.New("path not found")
	}
}
