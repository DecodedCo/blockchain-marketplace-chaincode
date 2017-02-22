
package utils


import (
    "net/http"
    "io/ioutil"
    "fmt"
    "encoding/json"
)


// ============================================================================================================================


// types and structs


// ============================================================================================================================


func GetJSONFromURL(url string) (interface{}, error) {
    
    var emptyInterface interface{}
    
    // Get the URL.
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return emptyInterface, err
    }
    defer resp.Body.Close()

    // Read the return
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return emptyInterface, err
    }

    // Marshal the data in an interface.
    var data interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        fmt.Println(err)
        return emptyInterface, err
    }

    return data, nil

}


// ============================================================================================================================

