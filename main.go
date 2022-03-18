// TODO - Create pod
// 1. Get pod annotation
//		a. check if it's marked as snowflake-id
//			TRUE: jump to (2. Verify ENV)
//			FALSE: return 0 byte and nil
//
// 2. Verify ENV
// 		a. No ENVs 			-> jump to (3. Get the replicaSet of the pod)
// 		b. Imcompleted ENVs	-> return nil and ERROR
// 		c. Intact ENVs 		-> return 0 byte and nil
//
// 3. Get the `replicaSet` of the pod
//		a. Retrieve the amount of pods which are controlled by its `replicaSet` (from Redis)
//		 	1. check if pods num <= 1024 (32 * 32)
//				TRUE: jump to (4. Schedule Pod to node)
//				FALSE: return nil and ERROR
//
// 4. Schedule Pod to node
//		a. Get available and ready worker nodes
//			1. Select a node
//			2. check if the node is registered
//				TRUE: Check if the `replicaSet` doesn't reach the maximum of 32 pods within the same worker node
//					TRUE: a. Add the pod to the node
// 						  b. return data center id and worker id (5. Attach ENV to pod)
//					FALSE: jump to (4.a.1)
//              FALSE: Check if the `replicaSet` doesn't exist on a maximum of 32 nodes
//					TRUE: a. Register the node and add the pod to it
// 						  b. return data center id and worker id (5. Attach ENV to pod)
//					FALSE: jump to (4.a.1)
//
// 5. Attach ENV to pod

package main

import (
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"flag"
	"strconv"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	parameters            = ServerParameters{}
)

func main() {
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	finish := make(chan bool)
	startServer()
	<-finish

	defer daprClient.Close()
}

func startServer() {
	go func() {
		healthHttp := http.NewServeMux()
		healthHttp.HandleFunc("/health", HandleHealth)
		log.Fatal(http.ListenAndServe(":43000", healthHttp))
	}()

	go func() {
		mutateHttp := http.NewServeMux()
		mutateHttp.HandleFunc("/mutate-v1-pod", HandlePodMutate)
		mutateHttp.HandleFunc("/validate-v1-pod", HandlePodValidate)
		log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, mutateHttp))
	}()
}
