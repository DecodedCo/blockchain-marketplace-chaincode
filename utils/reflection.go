

package utils


import (
    "reflect"
    "fmt"
)


// ============================================================================================================================


// types and structs


// ============================================================================================================================


func ReflectInterface(interFace interface{}, key string) interface{} {
    reflectedInterface := reflect.ValueOf(interFace)
    switch reflectedInterface.Kind() {
        // case reflect.Ptr:
            // fmt.Println("\t--- Pointer")
        // case reflect.Interface:
            // fmt.Println("\t--- Interface")
        // case reflect.Struct:
            // fmt.Println("\t--- Struct")
        // case reflect.String:
            // fmt.Println("\t--- String")
        // case reflect.Slice:
            // fmt.Println("\t--- Slice")
        case reflect.Map:
            // Iterate over all keys
            for _, k := range reflectedInterface.MapKeys() {
                originalValue := reflectedInterface.MapIndex(k)
                if k.String() == key {
                    // This has the key! So now we need to return the value as an interface
                    return originalValue.Interface()
                }
            }
        default:
            fmt.Println("\t--- Default")
    }
    return interFace
}


// EOF