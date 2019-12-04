package main

		import (
			"bytes"
			"fmt"
			"html/template"
			"io/ioutil"
			metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
			"k8s.io/client-go/kubernetes"
			"k8s.io/client-go/rest"
			"log"
			"os"
			"os/exec"
			"strings"
		)

		var config *rest.Config

		type Config1 struct {
			Matches []string
			Ipaddr  string
		    Container string
		    Username string
			Jobname string
		    Jobid string
		    Role string
		}

		var c Config1

		func getConfig() *rest.Config {
			var err error
			if config == nil {
				config, err = rest.InClusterConfig()
				if err != nil {
					panic(err.Error())
				}
			}
			return config
		}

		var clientset *kubernetes.Clientset

		func getClientset() *kubernetes.Clientset {

			var err error

			if clientset == nil {
				// creates the in-cluster config
				config := getConfig()

				// create the clientset
				clientset, err = kubernetes.NewForConfig(config)
				if err != nil {
					fmt.Println("Some error occured in creation of clientset:", err)
				}
			}
			return clientset
		}

		func Render(config Config1) {
			fmt.Println("\n called Render func to render ")

			var c bytes.Buffer

			template.Must(template.New("Template").Parse(`
		filebeat.inputs:
		- type: docker
		  containers:
		        ids:{{range $matchValue  :=  .Matches}}
		          - '{{$matchValue}}' {{end}}
		  fields:
		    Containerid: {{.Container}}
		    Jobid : {{.Jobid}}
		    Username: {{.Username}}
		    Jobname : {{.Jobname}}
		    Role : {{.Role}}
		    
		  fields_under_root: true

		output.logstash:
		        hosts: ["{{.Ipaddr}}:5044"]
		             
		     `)).Execute(&c, config)

			err := ioutil.WriteFile("/etc/filebeat/filebeat.yml", c.Bytes(), 0644)
			if err != nil {
				fmt.Println("error is ", err)
				return
			}

		}

		func execute() {
			out, err := exec.Command("/bin/sh", "-c", " service filebeat restart").Output()
			if err != nil {
				fmt.Printf("%s", err)
			}
			fmt.Println("Command Successfully Executed")
			output := string(out[:])
			fmt.Println(output)
		}

		func main() {
			fmt.Println("enter into main")
			// create the clientset
			clientset := getClientset()

			namespace := os.Getenv("MY_POD_NAMESPACE")
			podName := os.Getenv("MY_POD_NAME")
			ipaddr := os.Getenv("LOGSTASH_ENDPOINT")
		 
			pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
			if err != nil {
				log.Println("error is: ", err)
				return
			} else {
				for len(pod.Spec.Containers) != len(pod.Status.ContainerStatuses) {
					pod, err = clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
					if err != nil {
						log.Println("error is: ", err)
						return
					}
				}
				b := 0
				for b <= 1{
				    for _, containers := range pod.Status.ContainerStatuses {
					    if containers.Name != "filebeat" {
		                    a := strings.SplitAfter(containers.ContainerID, "docker://")
		                    b = len(a)
		                    if len(a) <= 1{
		                        break
		                    }else{
		                        containerID := a[1]
		                        c.Matches = append(c.Matches, containerID)
		                        c.Container = containerID
		                    }
		                }
		            }
		            pod, _ = clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
			    }
				c.Ipaddr = ipaddr
		        c.Jobid = os.Getenv("JOBUUID")
		        c.Username =os.Getenv("USERNAME")
		    	c.Jobname = os.Getenv("JOBNAME")
			 if os.Getenv("D3JOB_ROLE") !=  " "{
			 	c.Role =os.Getenv("D3JOB_ROLE")
			 }else{
			 	c.Role = " "
			 }
   		 	   out, err := exec.Command("/bin/sh", "-c", "./insert.sh").Output()
               if err != nil {
                 fmt.Printf("%s", err)
                }
            fmt.Println("Command Successfully Executed")
				log.Println("Rendering completed")
				for {
					pod, _ = clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
					for _, containers := range pod.Status.ContainerStatuses {
						if containers.Name != "filebeat" && containers.State.Terminated.String() != "nil" {
								log.Printf("Name: %s, Reason: %s\n", containers.Name, containers.State.Terminated.Reason)
								if containers.State.Terminated.Reason == "Completed" {
									log.Println("Hello I'm gonna die\n")
									os.Exit(0)
								}
						}
					}

				}
			}
			out, err := exec.Command("/bin/sh", "-c", "./delete.sh").Output()
            if err != nil {
                fmt.Printf("%s", err)
            }
            fmt.Println("Command Successfully Executed")
            output := string(out[:])
            fmt.Println(output)

		}

