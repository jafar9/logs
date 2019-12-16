package main

import (
        "log"
        "os"
)
  

func main() {
        log.Println("this is testing deployement")
        log.Println("env varibale print", os.Getenv("DEPLOYMENT_CONTEXT"))
}
 
