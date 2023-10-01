package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

var config *rest.Config
var clientSet *kubernetes.Clientset

type AppScheduler struct {
	pod *coreV1.Pod
}

const START_ID int = 1
const MAX_ID int = 32
const MAX_POD_NUM int = MAX_ID * MAX_ID

func (s *AppScheduler) schedulePod(podName string) (*snowflake, error) {
	// 3. Get the `replicaSet` of the pod
	rsName := s.pod.OwnerReferences[0].Name

	pa := populateApp(rsName)

	if getTotalPod(pa) == 961 {
		return &snowflake{}, fmt.Errorf("the serivce `%v` has reached the maximum pod numbers of %v", rsName, MAX_POD_NUM)
	}

	// 4. Schedule Pod to node
	nList, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.unschedulable=false",
	})
	if err != nil {
		log.Fatalf("Try to retrieve Nodes from K8s. Err: %s", err)
	}
	nodes := nList.Items

	for {
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)
		selectedIndex := r.Intn(len(nodes))
		selectedNodeName := nodes[selectedIndex].Name
		registered := false

		for i, n := range pa.Nodes {
			if n.Name == selectedNodeName { // Node is registered
				registered = true
				if len(n.Pods) == MAX_ID { // reach the maximum number of pods that a single node can hold
					break
				}

				// Add the pod to the node
				selectedID := -1

				sort.Sort(n.Pods)

				for id := START_ID; id < MAX_ID; id++ {
					selectedID = id
					if id > len(n.Pods)-1 {
						break
					}
					if id != n.Pods[id].ID {
						break
					}
				}

				pa.Nodes[i].Pods = append(pa.Nodes[i].Pods, pod{
					ID: selectedID,
				})

				pa.saveApp()

				return &snowflake{
					nodeName:     n.Name,
					datacenterId: n.ID,
					workerId:     selectedID,
				}, nil
			}
		}

		if !registered { // Node is unregistered
			if len(pa.Nodes) == MAX_ID { // reach the maximum number of nodes that app can scale on
				continue
			}

			selectedID := -1

			sort.Sort(pa.Nodes)

			for id := START_ID; id < MAX_ID; id++ {
				selectedID = id
				if id > len(pa.Nodes)-1 {
					break
				}
				if id != pa.Nodes[id].ID {
					break
				}
			}

			pa.Nodes = append(pa.Nodes, node{
				Name: selectedNodeName,
				ID:   selectedID,
				Pods: []pod{
					{
						ID: START_ID,
					},
				},
			})

			pa.saveApp()

			return &snowflake{
				nodeName:     selectedNodeName,
				datacenterId: selectedID,
				workerId:     START_ID,
			}, nil
		}
	}
}

func getTotalPod(a *App) int {
	var num int
	for _, no := range a.Nodes {
		num += len(no.Pods)
	}
	return num
}

func init() {
	c, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	config = c

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = cs
}
