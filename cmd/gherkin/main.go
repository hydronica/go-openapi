package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v27"
	messages "github.com/cucumber/messages/go/v22"
	"github.com/hydronica/go-config"

	"github.com/hydronica/go-openapi"
)

var debug bool

type conf struct {
	In      string `flag:"in" desc:"file/dir which contains gherkin.feature files"`
	Recurse bool   `flag:"r" comment:"recurse through all directories"`

	Out  string `flag:"out" comment:"generated openAPI file"`
	Base string `flag:"base" comment:"base openAPI file"`

	Title       string `flag:"-" comment:"title for openAPI doc"`
	Version     string `flag:"-" comment:"version of app for openAPI doc"`
	Description string `flag:"-" comment:"description for openAPI doc"`
}

func (c conf) Validate() error {
	if c.In == "" {
		return errors.New("input file/dir is required")
	}
	return nil
}

// Go
// Download the package via: `go get github.com/cucumber/cucumber/gherkin/go`
func main() {
	c := conf{
		Out:         "swag.json",
		Title:       "my app",
		Version:     "v0.10.14",
		Description: "describe me",
	}
	flag.BoolVar(&debug, "d", false, "show debug logs")
	config.LoadOrDie(&c)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create openAPI/Swagger doc
	var doc *openapi.OpenAPI
	if c.Base != "" {
		f, err := os.Open(c.Base)
		if err != nil {
			log.Fatalf("error reading base file %q: %v", c.Base, err)
		}
		b, err := io.ReadAll(f)
		if err != nil {
			log.Fatalf("error reading base file %q: %v", c.Base, err)
		}

		doc, err = openapi.NewFromJson(string(b))
		if err != nil {

		}
	} else {
		doc = openapi.New(c.Title, c.Version, c.Description)
	}

	//read and process gherkin files
	files, err := listFiles(c.In, c.Recurse)
	if err != nil {
		log.Fatal(err)
	}
	tests := make(routes)
	for _, f := range files {
		fileContent, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read file %q err: %v", f, err)
		}
		reader := strings.NewReader(string(fileContent))
		gherkinDocument, err := gherkin.ParseGherkinDocument(reader, newID)
		if err != nil {
			log.Fatal(err)
		}
		r := extractTest(gherkinDocument)
		tests.addRoutes(r)
	}

	// convert gherkin docs to routes
	for k, examples := range tests {
		s := strings.Split(k, "|")
		path, method := s[0], s[1]
		route, _ := doc.AddRoute(path, method, "", "", "")
		reqExamples := make([]openapi.ExampleObject, 0)
		for _, ex := range examples {

			if ex.ReqBody != "" {
				reqExamples = append(reqExamples, openapi.ExampleObject{
					Name:    ex.Name,
					Desc:    ex.Description,
					Example: rawJSON(ex.ReqBody),
				})
			}
			respExamples := []openapi.ExampleObject{}
			if ex.RespBody != "" {
				respExamples = append(respExamples, openapi.ExampleObject{Example: rawJSON(ex.RespBody)})

			}
			doc.AddResponse(route, openapi.NewRespBody("application/json", openapi.Code(ex.Status), "", respExamples))
		}
		doc.AddRequest(route, openapi.NewReqBody("application/json", "", reqExamples))

	}

	// generate the output swagger doc
	f, err := os.Create(c.Out)
	if err != nil {
		log.Fatalf("issue with writing %q: %w", c.Out, err)
	}
	f.Write(doc.JSON())
}

var counter int

func newID() string {
	counter += 1
	return "UUID" + strconv.Itoa(counter)
}

type test struct {
	Name        string
	Header      map[string]string
	Description string
	Method      string // POST/GET
	Path        string
	Body        string

	StatusCode string
	RespBody   string
}

var regURL = regexp.MustCompile(".*(POST|GET|PUT|DELETE).*\\\"(.*)\\\"")

func extractTest(document *messages.GherkinDocument) routes {
	tests := make(routes)
	for _, child := range document.Feature.Children {
		ex := Example{}
		if child.Scenario != nil {
			ex.Name = child.Scenario.Name
			ex.Description = child.Scenario.Description
			for _, step := range child.Scenario.Steps {
				switch step.KeywordType {
				case "Context", "Conjunction":
					switch step.Text {
					case "body of request:":
						ex.ReqBody = step.DocString.Content
					case "JSON response should be:":
						ex.RespBody = step.DocString.Content
					case "request headers:":
						ex.Header = processDataTable(step.DataTable)
					default:
						if debug {
							log.Printf("Unknown Text: %v", step.Text)
						}
					}
				case "Action":
					if !regURL.MatchString(step.Text) {
						log.Println("match not found:", step.Text)
						continue
					}
					m := regURL.FindStringSubmatch(step.Text)
					ex.method = strings.ToLower(m[1])
					ex.path = m[2]
				case "Outcome":
					if after, found := strings.CutPrefix(step.Text, "The status code should be "); found {
						i, err := strconv.Atoi(after)
						if err != nil && debug {
							log.Printf("unknown status code %q", after)
							continue
						}
						ex.Status = i
					}
				default:
					if debug {
						log.Printf("unknown keywordType: %v", step.KeywordType)
					}
				}
			}
			tests.AddExample(ex.path, ex.method, ex)
		}
	}

	return tests
}

func listFiles(path string, recurse bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat err: %q %w", path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	//list all files in directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}
	files := make([]string, 0)
	for _, f := range entries {
		if f.IsDir() && recurse {
			fmt.Println("dir", path, f.Name())
			f, err := listFiles(f.Name(), recurse)
			if err != nil {
				return nil, err
			}
			files = append(files, f...)
		}
		//add all files with the .feature extension
		if filepath.Ext(f.Name()) == ".feature" {
			files = append(files, path+"/"+f.Name())
		}

	}

	return files, nil
}

func processDataTable(data *messages.DataTable) map[string]string {
	m := make(map[string]string)
	isHeader := true
	for _, r := range data.Rows {
		if len(r.Cells) != 2 {
			log.Println("datatable: ", *data)
			return map[string]string{}
		}
		if isHeader && (r.Cells[0].Value == "key" && r.Cells[1].Value == "value") {
			isHeader = false
			continue
		}
		m[r.Cells[0].Value] = r.Cells[1].Value
	}
	return m
}

type routes map[string][]Example // [path|method][]Example

func (r routes) AddExample(path, method string, ex ...Example) {
	key := path + "|" + method
	examples, found := r[key]
	if !found {
		examples = make([]Example, 0)
	}
	r[key] = append(examples, ex...)
}

func (r routes) addRoutes(new routes) {
	for k, ex := range new {
		s := strings.Split(k, "|")
		path := s[0]
		method := s[1]
		r.AddExample(path, method, ex...)

	}
}

type Example struct {
	path   string
	method string

	Name        string
	Description string
	Header      map[string]string
	ReqBody     string

	Status   int
	RespBody string
}

func rawJSON(data string) map[string]any {
	m := make(map[string]any)
	err := json.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatal(err)
	}

	return m
}
