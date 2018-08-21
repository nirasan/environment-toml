package toml

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"go/ast"
	"reflect"
	"time"
	"unicode"
)

var nilValue = reflect.ValueOf(nil)

func Load(v interface{}, file string, env string) error {
	tree, err := toml.LoadFile(file)
	if err != nil {
		return err
	}

	if v == nil {
		return fmt.Errorf("v must not be nil")
	}
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
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

func castValue(t reflect.Type, v reflect.Value) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := v.Interface().(int64)
		rv := reflect.New(t).Elem()
		if rv.OverflowInt(n) {
			return nilValue, errors.New(fmt.Sprintf("%s is overflow: %v", t.Name, v))
		}
		rv.SetInt(n)
		return rv, nil
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := v.Interface().(int64)
		rv := reflect.New(t).Elem()
		if n < 0 {
			return nilValue, errors.New(fmt.Sprintf("%s is overflow: %v", t.Name, v))
		}
		if rv.OverflowUint(uint64(n)) {
			return nilValue, errors.New(fmt.Sprintf("%s is overflow: %v", t.Name, v))
		}
		rv.SetUint(uint64(n))
		return rv, nil
	case reflect.Float32, reflect.Float64:
		n := v.Interface().(float64)
		rv := reflect.New(t).Elem()
		if rv.OverflowFloat(n) {
			return nilValue, errors.New(fmt.Sprintf("%s is overflow: %v", t.Name, v))
		}
		rv.SetFloat(n)
		return rv, nil
	default:
		return v, nil
	}
}

func getValue(t reflect.Type, tree *toml.Tree, elem, env string) (reflect.Value, error) {
	switch {
	case t == reflect.TypeOf(time.Time{}):
		return getBasicValue(t, tree, elem, env)
	case t.Kind() == reflect.Struct:
		return getStructValue(t, tree, elem, env)
	case t.Kind() == reflect.Map:
		return getMapValue(t, tree, elem, env)
	case t.Kind() == reflect.Array, t.Kind() == reflect.Slice:
		return getArrayValue(t, tree, elem, env)
	default:
		v, e := getBasicValue(t, tree, elem, env)
		if e == nil {
			return castValue(t, v)
		}
		return v, e
	}
}

func getBasicValue(t reflect.Type, tree *toml.Tree, elem, env string) (reflect.Value, error) {
	p, err := findPath(tree, elem, env)
	if err != nil {
		return nilValue, err
	}
	v := tree.Get(p)
	vt := reflect.TypeOf(v)
	if vt.Kind() == reflect.Int64 {
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			//nop
		default:
			return nilValue, errors.New(fmt.Sprint("invalid type:", t, vt))
		}
	} else if vt.Kind() == reflect.Float64 {
		switch t.Kind() {
		case reflect.Float32, reflect.Float64:
			//nop
		default:
			return nilValue, errors.New(fmt.Sprint("invalid type:", t, vt))
		}
	} else if reflect.TypeOf(v) != t {
		return nilValue, errors.New("invalid type")
	}
	return reflect.ValueOf(v), nil
}

func getArrayValue(t reflect.Type, tree *toml.Tree, elem, env string) (reflect.Value, error) {
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return nilValue, errors.New("invalid type")
	}
	p, err := findPath(tree, elem, env)
	if err != nil {
		return nilValue, err
	}
	v := tree.Get(p)
	et := t.Elem()
	rv := reflect.MakeSlice(t, 0, 0)

	switch ary := v.(type) {
	case []*toml.Tree:
		for _, childTree := range ary {
			ev, e := getValue(et, childTree, "", elem)
			if e != nil {
				return nilValue, e
			}
			rv = reflect.Append(rv, ev)
		}
	case []interface{}:
		for _, a := range ary {
			av, e := castValue(et, reflect.ValueOf(a))
			if e != nil {
				return nilValue, errors.New("Invalid type")
			}
			rv = reflect.Append(rv, av)
		}
	default:
		return nilValue, errors.New("Invalid type")
	}
	return rv, nil
}

func getMapValue(t reflect.Type, tree *toml.Tree, elem, env string) (reflect.Value, error) {
	if t.Kind() != reflect.Map || t.Key().Kind() != reflect.String {
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
		if val, ok := newTree.(*toml.Tree); ok {
			target = val
		} else {
			return nilValue, errors.New("invalud tree")
		}
	}
	// get map value from tree
	rv := reflect.MakeMap(t)
	for _, k := range target.Keys() {
		v, err := getValue(t.Elem(), target, k, env)
		if err != nil {
			continue
		}
		rv.SetMapIndex(reflect.ValueOf(k), v)
	}
	return rv, nil
}

func getStructValue(t reflect.Type, tree *toml.Tree, elem, env string) (reflect.Value, error) {
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
		if val, ok := newTree.(*toml.Tree); ok {
			target = val
		} else {
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

func findPath(tree *toml.Tree, elem, env string) (string, error) {
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
