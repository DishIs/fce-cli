package display

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// Output Structured output for JSON or CSV formats.
// Returns true if output was handled (format was json or csv).
func Output(data interface{}) bool {
	if GlobalFormat == "text" {
		return false
	}

	if GlobalFormat == "json" {
		out, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			Error(fmt.Sprintf("Failed to marshal JSON: %v", err))
			return true
		}
		fmt.Println(string(out))
		return true
	}

	if GlobalFormat == "csv" {
		err := printCSV(data)
		if err != nil {
			Error(fmt.Sprintf("Failed to output CSV: %v", err))
		}
		return true
	}

	return false
}

func printCSV(data interface{}) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return nil // empty csv
		}
		// Write headers based on first element
		first := v.Index(0)
		if first.Kind() == reflect.Interface {
			first = first.Elem()
		}
		
		if first.Kind() == reflect.Map {
			var headers []string
			for _, key := range first.MapKeys() {
				headers = append(headers, fmt.Sprintf("%v", key.Interface()))
			}
			w.Write(headers)
			
			// Write rows
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i)
				if item.Kind() == reflect.Interface {
					item = item.Elem()
				}
				var row []string
				for _, key := range first.MapKeys() {
					val := item.MapIndex(key)
					if val.IsValid() {
						row = append(row, fmt.Sprintf("%v", val.Interface()))
					} else {
						row = append(row, "")
					}
				}
				w.Write(row)
			}
			return nil
		} else if first.Kind() == reflect.String {
			w.Write([]string{"value"})
			for i := 0; i < v.Len(); i++ {
				w.Write([]string{fmt.Sprintf("%v", v.Index(i).Interface())})
			}
			return nil
		}
	} else if v.Kind() == reflect.Map {
		// Single object
		var headers []string
		var row []string
		for _, key := range v.MapKeys() {
			headers = append(headers, fmt.Sprintf("%v", key.Interface()))
			row = append(row, fmt.Sprintf("%v", v.MapIndex(key).Interface()))
		}
		w.Write(headers)
		w.Write(row)
		return nil
	} else {
		// Single string or other
		w.Write([]string{"value"})
		w.Write([]string{fmt.Sprintf("%v", data)})
	}

	return nil
}
