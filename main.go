package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
)

var delimiter = flag.String("delimiter", "\t", "delimiter")
var showHelp = flag.Bool("help", false, "display help")
var runServer = flag.Bool("server", false, "view in browser instead")
var serverPort = flag.Int("port", 8888, "port for server")

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

	if *runServer {
		serveHTMLPage(*serverPort, res)
	} else {
		printToStdOut(res)
	}
}

func serveHTMLPage(port int, res []interface{}) {
	http.HandleFunc("/json", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(res)
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		pattern := `<html>
<head>
	<script src="https://unpkg.com/gridjs/dist/gridjs.production.min.js"></script>
	<link href="https://unpkg.com/gridjs/dist/theme/mermaid.min.css" rel="stylesheet" />
</head>
<body>
    <div id="wrapper"></div>
	<script>
	new gridjs.Grid({
	  search: true,
	  sort: true,
	  columns: {{.Headers}},
	  data: {{.Data}},
	}).render(document.getElementById("wrapper"));
	</script>
</body></html>`
		glob, err := template.New("").Parse(pattern)
		if err != nil {
			log.Fatal(err)
		}
		headers := extractHeaders(res)
		if err := glob.Execute(writer, map[string]interface{}{
			"Headers": headers,
			"Data":    transformDataForFrontend(res, headers),
		}); err != nil {
			log.Fatal(err)
		}
	})
	addr := fmt.Sprintf("localhost:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Visit http://%s\n", addr)
	if err := http.Serve(l, nil); err != nil {
		log.Fatal(err)
	}
}

// The frontend library doesn't handle nested objects well so we just make sure it's one layer deep.
func transformDataForFrontend(input []interface{}, headers []string) interface{} {
	var res []interface{}
	for _, re := range input {
		item, ok := re.(map[string]interface{})
		if !ok {
			res = append(res, re)
			continue
		}
		newItem := make(map[string]string)
		for _, header := range headers {
			newItem[header] = stringify(item[header])
		}
		res = append(res, newItem)
	}
	return res
}

func printToStdOut(res []interface{}) {
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
