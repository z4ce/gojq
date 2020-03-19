package structbuilder

import (
	"fmt"
	"reflect"

	"github.com/itchyny/gojq"
)

var doDebug = false

// For map requires you to match the types
func processMap(tobeFilled reflect.Value, queryIter gojq.Iter) error {
	// maybe we should check this instead of just crashing?
	jqResult, _ := queryIter.Next()
	presumedMap, ok := jqResult.(map[string]interface{})

	if !ok {
		return fmt.Errorf("jq query didn't return map in map context for %v. Got this instead: %v", tobeFilled, jqResult)
	}

	elemType := tobeFilled.Type().Elem().Kind()
	if elemType == reflect.Struct {
		newMap := reflect.MakeMap(tobeFilled.Type())
		for k, v := range presumedMap {
			printDebug("Processing kv %v %v\n", k, v)
			fo, err := FillOut(reflect.New(tobeFilled.Type().Elem()).Elem(), v)
			if err != nil {
				return err
			}
			printDebug("fo %v\n", fo)
			newMap.SetMapIndex(reflect.ValueOf(k), fo.(reflect.Value))
		}
		printDebug("Newmap %v\n", newMap)
		tobeFilled.Set(newMap)
	} else if elemType == reflect.Interface {
		tobeFilled.Set(reflect.ValueOf(presumedMap))
	} else {
		printDebug("Map: %v\n", tobeFilled)
		return fmt.Errorf("No processor type for %v", tobeFilled)
	}
	return nil
}

// the query must return a string.
// possible enhancement could be to turn whatever is selected into a string
func processString(tobeFilled reflect.Value, queryIter gojq.Iter) error {
	jqResult, _ := queryIter.Next()
	presumedStr, ok := jqResult.(string)

	if !ok {
		return fmt.Errorf("jq query didn't return string in string context for %v. Got this instead: %v", tobeFilled, jqResult)
	}

	tobeFilled.SetString(presumedStr)
	return nil
}

// convert from int to string
// the upstream treats all numbers as floats,
// for our purposes that seems dangerous for large values, but it seems to be working ok
// for now. json spec only lets you use 32bits, which is lossless for 64bit, so this /should/ be ok if you're in spec.
func processInt64(tobeFilled reflect.Value, queryIter gojq.Iter) error {
	jqResult, _ := queryIter.Next()
	presumedFloat, ok := jqResult.(float64)
	if !ok {
		return fmt.Errorf("jq query didn't return int or float in int context for %v. Got this instead: %v", tobeFilled, jqResult)
	}
	//this seems like there could be data loss.
	presumedInt := int64(presumedFloat)
	tobeFilled.SetInt(presumedInt)
	return nil
}

func processFloat64(tobeFilled reflect.Value, queryIter gojq.Iter) error {
	jqResult, _ := queryIter.Next()
	presumedFloat, ok := jqResult.(float64)
	if !ok {
		return fmt.Errorf("jq query didn't return float in int context for %v. Got this instead: %v", tobeFilled, jqResult)
	}
	tobeFilled.SetFloat(presumedFloat)
	return nil
}

func processBool(tobeFilled reflect.Value, queryIter gojq.Iter) error {
	jqResult, _ := queryIter.Next()
	presumedBool, ok := jqResult.(bool)
	if !ok {
		return fmt.Errorf("jq query didn't return bool in bool context for %v. Got this instead: %v", tobeFilled, jqResult)
	}
	tobeFilled.SetBool(presumedBool)
	return nil
}

// FillOut takes a struct that has been annotated with `jq` tags specifying
// what queries should run to fill in that field and the document that
// queries will be run against. The document should be of the format
// returned by json.Unmarshal().
// See the TestRoundTrip test case for a good example.
// Currently works with the following embedded types:
// - string
// - int64
// - float64
// - bool
// - map[string]interface{} - arbitrary json
// - map[]struct
// TODO:
// - struct
// - arrays
func FillOut(tobeFilled interface{}, doc interface{}) (interface{}, error) {
	printDebug("tobe %v\n", tobeFilled)
	var v reflect.Value

	if reflect.TypeOf(tobeFilled) == reflect.TypeOf(reflect.Value{}) {
		v = tobeFilled.(reflect.Value)
	} else {
		v = reflect.ValueOf(tobeFilled).Elem()
	}

	printDebug("begin for loop %v\n", v.NumField())
	for index := 0; index < v.NumField(); index++ {
		f := v.Field(index)
		printDebug("ftype %v", v.Type().Field(index))
		jq, ok := v.Type().Field(index).Tag.Lookup("jq")
		printDebug("for: %v\n", jq)
		if !ok {
			continue
		}
		query, err := gojq.Parse(jq)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse jq query for %s for field %v", jq, f)
		}
		queryIter := query.Run(doc)

		switch f.Type().Kind() {
		case reflect.Map:
			err = processMap(f, queryIter)
		case reflect.String:
			err = processString(f, queryIter)
		case reflect.Int64:
			err = processInt64(f, queryIter)
		case reflect.Float64:
			err = processFloat64(f, queryIter)
		case reflect.Bool:
			err = processBool(f, queryIter)
		}
		if err != nil {
			return nil, err
		}
		printDebug("f %v\n", f)

	}

	return tobeFilled, nil
}

func printDebug(str string, args ...interface{}) {
	if doDebug {
		fmt.Printf(str, args...)
	}
}
