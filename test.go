package main

import (
        "github.com/golang/protobuf/proto"
        "log"
        mlpb "ml_metadata/proto/metadata_store_go_proto"
        storepb "ml_metadata/proto/metadata_store_service_go_proto"

        "context"
        "flag"
        "fmt"
        "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
        wfClient "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
        "google.golang.org/grpc"
        "k8s.io/api/core/v1"
        "k8s.io/client-go/rest"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/apimachinery/pkg/labels"
        "k8s.io/client-go/tools/clientcmd"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/apimachinery/pkg/watch"
        kcache "k8s.io/client-go/tools/cache"
        "k8s.io/klog"
        "time"
        "os"
)

var (
        masterURL          string
        kubeconfig         string
        metadataServiceURL string
        resourcelist       string
)

const workspace = "dkube_runs"

type MetaLogger struct {
        // Metadata gRPC client.
        kfmdClient storepb.MetadataStoreServiceClient
        typeID     int64
        atype      string
}

var artifact_type *MetaLogger
var execution_type *MetaLogger

func mlpbStringValue(s string) *mlpb.Value {
        return &mlpb.Value{
                Value: &mlpb.Value_StringValue{
                        StringValue: s,
                },
        }
}

// NewMetaLogger creates a new MetaLogger for a specific k8s GroupVersionKind.
func ArtifactTypeCreation(kfmdClient storepb.MetadataStoreServiceClient) (*MetaLogger, error) {
        l := &MetaLogger{
                kfmdClient: kfmdClient,
                atype:      "dkube/jobs/dataset",
        }
        resourceArtifactType := mlpb.ArtifactType{
                Name: proto.String(l.atype),
                Properties: map[string]mlpb.PropertyType{
                        // dataset name
                        "name": mlpb.PropertyType_STRING,
                        // dkube version
                        "version": mlpb.PropertyType_STRING,
                        // wf creation time
                        "create_time": mlpb.PropertyType_STRING,
                },
        }
        request := storepb.PutArtifactTypeRequest{
                ArtifactType:   &resourceArtifactType,
                AllFieldsMatch: proto.Bool(true),
        }
        resp, err := kfmdClient.PutArtifactType(context.Background(), &request)
        if err != nil {
                return l, fmt.Errorf("failed to create artifact type: err = %v; request = %v; response = %v", err, request, resp)
        }
        l.typeID = resp.GetTypeId()
        log.Println("typeid ", l.typeID)
        return l, nil
}

func ArtifactCreation(labels map[string]string, path string) (int64, error) {
        var id int64
        artifact := mlpb.Artifact{
                TypeId: proto.Int64(artifact_type.typeID),
                Uri:    proto.String(path),
                Properties: map[string]*mlpb.Value{
                        "name":        mlpbStringValue(labels["datasets"]),
                        "version":     mlpbStringValue("1.4.1"),
                        "create_time": mlpbStringValue(time.Now().Format(time.RFC3339)),
                },
                CustomProperties: map[string]*mlpb.Value{
                        // set the workspace to group the metadata.
                        "__kf_workspace__": mlpbStringValue(workspace),
                        "__kf_run__":       mlpbStringValue(labels["jobname"]),
                },
        }
        request := storepb.PutArtifactsRequest{
                Artifacts: []*mlpb.Artifact{&artifact},
        }
        resp, err := artifact_type.kfmdClient.PutArtifacts(context.Background(), &request)
        if err != nil {
                klog.Errorf("failed to log metadata for %s: err = %s, request = %v, resp = %v", artifact_type.atype, err, request, resp)
                return id, err
        }
        klog.Infof("Handled addEvent for %s.\n", artifact_type.atype)
        return resp.ArtifactIds[0], nil
}

// NewMetaLogger creates a new MetaLogger for a specific k8s GroupVersionKind.
func ExecutionTypeCreation(kfmdClient storepb.MetadataStoreServiceClient) (*MetaLogger, error) {
        l := &MetaLogger{
                kfmdClient: kfmdClient,
                atype:      "dkube/jobs/executions",
        }
        resourceExecutionType := mlpb.ExecutionType{
                Name: proto.String(l.atype),
                Properties: map[string]mlpb.PropertyType{
                        // same as metav1.Object.Name
                        "jobname": mlpb.PropertyType_STRING,
                        "jobuuid": mlpb.PropertyType_STRING,
                        "jobid": mlpb.PropertyType_STRING,
                        "class": mlpb.PropertyType_STRING,
                },
        }
        request := storepb.PutExecutionTypeRequest{
                ExecutionType:  &resourceExecutionType,
                AllFieldsMatch: proto.Bool(true),
        }
        resp, err := kfmdClient.PutExecutionType(context.Background(), &request)
        if err != nil {
                return l, fmt.Errorf("failed to create artifact type: err = %v; request = %v; response = %v", err, request, resp)
        }
        l.typeID = resp.GetTypeId()
        log.Println("typeid  ", l.typeID)
        return l, nil
}

func ExecutionCreation(labels map[string]string) (int64, error) {
        var id int64
        execution := mlpb.Execution{
                TypeId: proto.Int64(execution_type.typeID),
                Properties: map[string]*mlpb.Value{
                        "jobname": mlpbStringValue(labels["jobname"]),
                        "jobuuid": mlpbStringValue(labels["jobuuid"]),
                        "jobid": mlpbStringValue(labels["jobid"]),
                        "class": mlpbStringValue(labels["class"]),
                },
                CustomProperties: map[string]*mlpb.Value{
                        // set the workspace to group the metadata.
                        "__kf_workspace__": mlpbStringValue(workspace),
                },
        }
        request := storepb.PutExecutionsRequest{
                Executions: []*mlpb.Execution{&execution},
        }
        resp, err := execution_type.kfmdClient.PutExecutions(context.Background(), &request)
        if err != nil {
                klog.Errorf("failed to log metadata for %s: err = %s, request = %v, resp = %v", execution_type.atype, err, request, resp)
                return id, err
        }
        klog.Infof("Handled addEvent for %s.\n", execution_type.atype)
        return resp.ExecutionIds[0], nil
}

func (l *MetaLogger) EventCreation(aids, eids int64) error {
        etype := mlpb.Event_DECLARED_INPUT
        event := mlpb.Event{
                ArtifactId:  &aids,
                ExecutionId: &eids,
                Type:        &etype,
        }
        request := storepb.PutEventsRequest{
                Events: []*mlpb.Event{&event},
        }
        resp, err := l.kfmdClient.PutEvents(context.Background(), &request)
        if err != nil {
                klog.Errorf("failed to log metadata for %s: err = %s, request = %v, resp = %v", l.atype, err, request, resp)
                return err
        }
        klog.Infof("Handled addEvent for %s.\n", l.atype)
        return nil
}

func main() {
        klog.InitFlags(nil)
        flag.Parse()

        // Set up a connection to the gRPC server.
        conn, err := grpc.Dial(metadataServiceURL, grpc.WithInsecure())
        if err != nil {
                klog.Fatalf("Faild to connect grpc server: %v", err)
        }

        kfmdClient := storepb.NewMetadataStoreServiceClient(conn)


        artifact_type, err = ArtifactTypeCreation(kfmdClient)
        if err != nil{
                log.Println(err)
        }

        log.Printf("%+v\n", artifact_type)
        execution_type, err = ExecutionTypeCreation(kfmdClient)
        if err != nil{
                log.Println(err)
        }
        log.Printf("%+v\n", execution_type)

        clientset := ArgoClientset()
        watchWfs(clientset)

        select {}
}

func wfAdded(obj interface{}) {
        wf := obj.(*v1alpha1.Workflow)

        //log.Printf("%+v\n", wf.ObjectMeta)

        log.Printf("artifact type %+v\n", artifact_type)
        path := wf.ObjectMeta.Annotations["sources"]

        aids, err := ArtifactCreation(wf.ObjectMeta.Labels, path)
        if err != nil {
                log.Println(err)
        }

        log.Printf("exee type %+v\n",execution_type)
        eids, err := ExecutionCreation(wf.ObjectMeta.Labels)
        if err != nil {
                log.Println(err)
        }
        log.Println("art id is", aids)
        log.Println("exe id is", eids)

        err = artifact_type.EventCreation(aids, eids)
        if err != nil {
                log.Println(err)
        }
}
/*
func wfUpdated(oldObj interface{}, newObj interface{}) {
        wf := newObj.(*v1alpha1.Workflow)
        //log.Printf("%+v\n", wf)
}

func wfDeleted(obj interface{}) {
        wf := obj.(*v1alpha1.Workflow)
        //log.Printf("%+v\n", wf)
}
*/
func watchWfs(client *wfClient.ArgoprojV1alpha1Client) {
        set := make(map[string]string)
        set["workflows.argoproj.io/controller-instanceid"] = "dkube"
        watchlist := NewListWatchFromClient(client.RESTClient(), "workflows", v1.NamespaceAll, labels.SelectorFromSet(set))
        resyncPeriod := 30 * time.Minute

        //Setup an informer to call functions when the watchlist changes
        _, eController := kcache.NewInformer(
                watchlist,
                &v1alpha1.Workflow{},
                resyncPeriod,
                kcache.ResourceEventHandlerFuncs{
                        //UpdateFunc: wfUpdated,
                        //DeleteFunc: wfDeleted,
                        AddFunc:    wfAdded,
                },
        )

        //Run the controller as a goroutine
        go eController.Run(wait.NeverStop)
}

func GetConfig() (*rest.Config, error) {
        if os.Getenv("KUBERNETES_PORT") == "" {
                var kubeConfigPath string
                if configPath := os.Getenv("KUBECONFIG"); configPath != "" {
                        kubeConfigPath = configPath
                } else {
                        homeDir := os.Getenv("HOME")
                        kubeConfigPath = homeDir + "/.kube/config"
                }
                return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
        } else {
                // creates the in-cluster config
                return rest.InClusterConfig()
        }
}

func ArgoClientset() *wfClient.ArgoprojV1alpha1Client {
        var config *rest.Config
        var err error
        var argoClientset *wfClient.ArgoprojV1alpha1Client

        if argoClientset == nil {
                config, err = GetConfig()
                if err != nil {
                        log.Println("StudyjobClientset() failed - config not found", err)
                        panic(err)
                }

                argoClientset, err = wfClient.NewForConfig(config)
                if err != nil {
                        log.Println("StudyjobClientset() failed ", err)
                        panic(err)
                }
        }
        return argoClientset
}

func NewListWatchFromClient(c kcache.Getter, resource string, namespace string, labelSelector labels.Selector) *kcache.ListWatch {
        listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
                options.LabelSelector = labelSelector.String()
                return c.Get().
                        Namespace(namespace).
                        Resource(resource).
                        VersionedParams(&options, metav1.ParameterCodec).
                        Do().
                        Get()
        }
        watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
                options.Watch = true
                options.LabelSelector = labelSelector.String()
                return c.Get().
                        Namespace(namespace).
                        Resource(resource).
                        VersionedParams(&options, metav1.ParameterCodec).
                        Watch()
        }
        return &kcache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

func init() {
        flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
        flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
        flag.StringVar(&metadataServiceURL, "metadata_service", "", "The address of the Kubeflow Metadata GRPC service. Required.")
        flag.StringVar(&resourcelist, "resourcelist", "", "The path of a JSON file with a list of Kubernetes GroupVersionKind to be watched. Required.")
}

