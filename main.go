package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var delimiter = flag.String("delimiter", "\t", "delimiter")
var showHelp = flag.Bool("help", false, "display help")

func main() {
	flag.Parse()

	if *showHelp {
		printHelpMessage()
		os.Exit(1)
	}

	decoder := json.NewDecoder(os.Stdin)
	var res []interface{}
	err := decoder.Decode(&res)
	if err != nil {
		fmt.Printf("cannot decode input: %w\n", err)
		os.Exit(1)
	}

	headers := extractHeaders(res)

	fmt.Println(strings.Join(headers, *delimiter))

	for _, re := range res {
		item, ok := re.(map[string]interface{})
		if !ok {
			continue
		}
		var row []string
		for _, header := range headers {
			row = append(row, stringify(item[header]))
		}
		fmt.Println(strings.Join(row, *delimiter))
	}
}

func printHelpMessage() {
	fmt.Println(`Example usage: 
Windows Powershell: 
>> echo '[{"foo":1.4, "bar": true},{"foo": 2.2},{"foo": 3}]' | go-json-table | column -t -s "` + "`t" + `"
foo       bar
1.400000  true
2.200000  -
3.000000  -

Linux:
$ echo '[{"foo":1.4, "bar": true},{"foo": 2.2},{"foo": 3}]' | go-json-table | column -t
foo       bar
1.400000  true
2.200000  -
3.000000  -

Using a different delimiter
$ echo '[{"foo":1.4, "bar": true},{"foo": 2.2},{"foo": 3}]' | go-json-table -delimiter "," | column -t -s ","
foo       bar
1.400000  true
2.200000  -
3.000000  -`)
}

func extractHeaders(res []interface{}) []string {
	var headers []string
	for _, re := range res {
		item, ok := re.(map[string]interface{})
		if !ok {
			continue
		}
		for s := range item {
			headers = append(headers, s)
		}
		break
	}
	return headers
}

func stringify(x interface{}) string {
	s, ok := x.(fmt.Stringer)
	if ok {
		return s.String()
	}
	if x == nil {
		return "-"
	}
	switch t := reflect.TypeOf(x).Name(); t {
	case "float64", "float32":
		return fmt.Sprintf("%f", x)
	case "int64", "int", "int32":
		return fmt.Sprintf("%d", x)
	case "string":
		return x.(string)
	default:
		d, _ := json.Marshal(x)
		return string(d)
	}
}
