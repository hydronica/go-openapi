# GoOpenAPI
A Go Lang SDK to help create OpenApi 3.0.3 Spec 

[OpenAPI spec](https://swagger.io/specification/)
## Getting Started 

``` go 
import (
    _ "embed"
    
    "github.com/hydronica/go-openapi"
)

// go:embed base.json 
var base string 
func main() {

    // create doc from base template
    doc, err := openapi.NewFromJson(base)
    if err != nil {
        log.Fatal(err) 
    }
    
    // create doc from scratch
    doc = openapi.New("title", "v1.0.0", "all about this API") 
   
    doc.AddRoute(
        openapi.NewRoute("/path/v1", "GET").
            AddResponse(
                openapi.Resp{Code: 200, Desc:"valid response"}.WithJSONString('{"status":"ok"}'
                ). 
            AddRequest(
                openapi.Req{MType: "application/json", Desc:"pull data"}.
                    WithParams(myStruct)
                )
    ) 
   
   // print generated json document
   fmt.Println(string(doc.JSON()))
}
```


### Overview 
 <img src="docs/chart.drawio.svg">
