
package utils


import (
    "net/http"
    "io/ioutil"
    "fmt"
    "reflect"
    "strconv"

    "crypto/tls"
    "crypto/x509"
)


// ============================================================================================================================


// types and structs


// ============================================================================================================================


func GetEndpoint(url string, keys []string) (string) {
    
    var data interface{}
    
    // Grab the interface from the endpoint.
    data, err := GetJSONFromURL(url)
    if err != nil {
        fmt.Println(err)
    }
    
    // Recursive reflection of the json interface.
    for _, key := range keys {
        data = ReflectInterface(data, key)
    }
    
    var output string
    switch data.(type) {
        case float64:
            output = strconv.FormatFloat(reflect.ValueOf(data).Float(), 'f', 6, 64)
        default:
            output = reflect.ValueOf(data).String()
    }
    
    return output

} // end of GetEndpoint


func GetExchangeRate() {
    
    dataInterface, err := GetJSONFromURL("http://api.fixer.io/latest?base=USD&symbols=GBP")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Printf("Results: %v\n", dataInterface)

} // end of GetExchangeRate


func GetBTCUSD() {
    
    dataInterface, err := GetJSONFromURL("http://api.coindesk.com/v1/bpi/currentprice.json")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Printf("Results: %v\n", dataInterface)

} // end of GetBTCUSD


// ============================================================================================================================

