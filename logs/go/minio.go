package main

import(
	"fmt"
	"bytes"
	"log"
	"io/ioutil"
	"io"
	"encoding/json"
	"net/http"
	"time"
	"net"
)

type LoginModel struct{
    Id int  `json:"id"`
    Jsonrpc string `json:"jsonrpc"`
    Method string `json:"method"`
    Params LoginCred `json:"params"`
}

type LoginCred struct{
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginRes struct{
    Id int  `json:"id"`
    Jsonrpc string `json:"jsonrpc"`
    Result LoginResult `json:"result"`
}

type LoginResult struct{
    Token string   `json:"token"`
    UiVersion string  `json:"uiVersion"`

}

type ObjectDetails struct{
    BucketName string `json:"bucketName"`
    Objects  []string `json:"objects"`
    Prefix string  `json:"prefix"`
}


func main(){
     http.HandleFunc("/logs", d3apiLogs)
    fmt.Println("the server is listening ....192.168.200.23:9401")
     err := http.ListenAndServe("192.168.200.23:9401", nil)

        if err != nil {
                fmt.Println(err)
        }
}

func d3apiLogs(w http.ResponseWriter, r *http.Request){
    login := LoginCred{Username: "dkube", Password: "l06dands19s"}
    jsonValue, _ := json.Marshal(&LoginModel{Id: 1, Jsonrpc: "2.0", Method: "Web.Login", Params: login})
    res, err := InvokeD3API("http://192.168.200.23:32223/minio/webrpc", "",jsonValue, "POST", false)
    if err != nil {
	    fmt.Println(err)
        return
    }
    if res.StatusCode != 200 {
            rdata, _ := ioutil.ReadAll(res.Body)
            log.Printf("submitting dkube job failed, reason - %s \n", string(rdata))
            return
        }
        responseBody, err := ioutil.ReadAll(res.Body)
        if err != nil{
            log.Println(err)
        }
        loginres := LoginRes{}
        err = json.Unmarshal([]byte(responseBody), &loginres)
		if err != nil {
		    log.Println(err)
			return 
		}
		fmt.Println("----->",loginres.Result.Token)
		if loginres.Result.Token != ""{
		    jsonValue, _ := json.Marshal(&LoginModel{Id: 1, Jsonrpc: "2.0", Method: "Web.CreateURLToken"})
		    token := "Bearer "+ loginres.Result.Token
		    res, err := InvokeD3API("http://192.168.200.23:32223/minio/webrpc", token, jsonValue, "POST", false)
		    if err != nil {
                fmt.Println(err)
                return
            }
            if res.StatusCode != 200 {
                    rdata, _ := ioutil.ReadAll(res.Body)
                    log.Printf("submitting dkube job failed, reason - %s \n", string(rdata))
                    return
                }
                responseBody, err := ioutil.ReadAll(res.Body)
                if err != nil{
                    log.Println(err)
                }
                loginres = LoginRes{}
                err = json.Unmarshal([]byte(responseBody), &loginres)
                if err != nil {
                    log.Println(err)
                    return
                }
                fmt.Println(loginres.Result.Token)

                objects := []string{"ac4c0082-e690-4259-90f7-0de6c2fb1c0c/"}
                jsonValue, _ = json.Marshal(&ObjectDetails{BucketName: "dkube", Prefix: "system/logs/sarathdasari/", Objects: objects})

                url := "http://192.168.200.23:32223/minio/zip"
                request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
                tr := &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout:   60 * time.Second,
                    KeepAlive: 60 * time.Second,
                }).DialContext,
                }
                client := &http.Client{Transport: tr, Timeout: time.Second * 60}
                    request.Header.Set("Accept-Encoding", "gzip, deflate")
                    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
                    q := request.URL.Query()
                    q.Add("token", loginres.Result.Token)
                    request.URL.RawQuery = q.Encode()
                    fmt.Println(request.URL.String())
                response, err := client.Do(request)

                w.Header().Set("Content-Disposition", "attachment; filename=test.zip")
	            w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	            fmt.Printf("%+v", response)

                 _, err = io.Copy(w, response.Body)
                 if err != nil{
                    log.Println(err)
                 }
                 return
            }

    return
}

func InvokeD3API(url string, token string, data []byte, method string, zipheaders bool) (*http.Response, error) {
    request, _ := http.NewRequest(method, url, bytes.NewBuffer(data))
    request.Header.Set("Content-Type", " application/json; charset=utf-8")

    tr := &http.Transport{
	//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}).DialContext,
    }
    client := &http.Client{Transport: tr, Timeout: time.Second * 60}
    if zipheaders == true{
        request.Header.Set("Accept-Encoding", "gzip, deflate")
        request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
        q := request.URL.Query()
        q.Add("token", token)
        request.URL.RawQuery = q.Encode()
        fmt.Println(request.URL.String())
    } else if token != ""{
        request.Header.Set("Authorization", token)
    }
    response, err := client.Do(request)
    return response, err
}

/*package main

import (
	"github.com/minio/minio-go"
	"log"
)

func main() {
	endpoint := "192.168.200.23:32223"
	accessKeyID := "dkube"
	secretAccessKey := "l06dands19s"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%+v\n", minioClient) // minioClient is now setup
}

package main
import(
    "github.com/aws/aws-sdk-go/awk"
    "fmt"
    "os"
    "path/filepath"
)

func DownloadFromS3Bucket(bucket, item, path string) {
    file, err := os.Create(filepath.Join(path, item))
    if err != nil {
        fmt.Printf("Error in downloading from file: %v \n", err)
        os.Exit(1)
    }

    defer file.Close()

    miniocred := &aws.Credentials.Value{AccessKeyID: "dkube", SecretAccessKey: "l06dands19s"}

    cred := &aws.Credentials.Credentials{creds: miniocred}

    sess, _ := session.NewSession(&aws.Config{
        Region: aws.String(constants.AWS_REGION), Credentials: cred, Endpoint: "192.168.200.23:32223", DisableSSL: "true"},
    )

    // Create a downloader with the session and custom options
    downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
        d.PartSize = 64 * 1024 * 1024 // 64MB per part
        d.Concurrency = 6
    })

    numBytes, err := downloader.Download(file,
        &s3.GetObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(item),
        })
    if err != nil {
        fmt.Printf("Error in downloading from file: %v \n", err)
        os.Exit(1)
    }

    fmt.Println("Download completed", file.Name(), numBytes, "bytes")
}*/

