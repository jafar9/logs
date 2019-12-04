package main
import (
    "log"
   "io/ioutil"
    "os"
)


func main(){

f, err := os.Open("num2.txt")
if err != nil{
        log.Println("File opening ERROR", err)
        return
}
 o2, err := f.Seek(-11240, 2)
 if err != nil{
         o2, err = f.Seek(0,1)
}
log.Println("old seek posotion is", o2)
 data, err := ioutil.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Data as string: %s\n", data)
    log.Println("Number of bytes read:", len(data))
}
