package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	coreV1 "k8s.io/api/core/v1"
)

var (
	mu sync.Mutex
)

func HandlePod(ad *admission) ([]byte, error) {
	// 1. Get pod annotation
	op := ad.review.Request.Operation
	pod := &coreV1.Pod{}

	var raw []byte

	if op == "CREATE" {
		raw = ad.review.Request.Object.Raw
	} else {
		raw = ad.review.Request.OldObject.Raw
	}

	if err := json.Unmarshal(raw, pod); err != nil {
		fmt.Errorf("could not unmarshal pod on admission request: %v", err)
	}

	if enabledSF := isSnowflakeApp(pod); !enabledSF {
		return []byte{}, nil
	}

	if op == "CREATE" {
		return addEnvToPod(ad, pod)
	} else if op == "DELETE" {
		go removePod(ad, pod)
		return []byte{}, nil
	} else { // Not in the case
		return nil, nil
	}
}

func isSnowflakeApp(p *coreV1.Pod) bool {
	for key, value := range p.Annotations {
		enabled, _ := strconv.ParseBool(value)
		if key == "snowflake-id.io/enabled" && enabled {
			return true
		}
	}
	return false
}

func addEnvToPod(ad *admission, pod *coreV1.Pod) ([]byte, error) {
	// 2. Verify ENV
	envs := pod.Spec.Containers[0].Env
	var dataCenterIDIsSet bool
	var workerIDIsSet bool

	// Avoid duplicate ENV fields
	for _, env := range envs {
		if env.Name == "SNOWFLAKE_DATA_CENTER_ID" {
			dataCenterIDIsSet = true
		} else if env.Name == "SNOWFLAKE_WORKER_ID" {
			workerIDIsSet = true
		}
	}

	if dataCenterIDIsSet && workerIDIsSet { // 2.c
		return []byte{}, nil
	}

	if !dataCenterIDIsSet && !workerIDIsSet { // 2.a
		// 3. Get the `replicaSet` of the pod

		// ---WARNING---
		// Mutex only works if there is only one copy of controller itself.
		//
		// TODO - The Architecture should be refactored to ditributed lock.
		// ---WARNING---
		mu.Lock()
		s := AppScheduler{
			pod: pod,
		}

		sf, err := s.schedulePod(ad.review.Request.Name)
		log.Printf("scheduled pod is %+v", *sf)

		if err != nil {
			return nil, err
		}
		mu.Unlock()

		patches := ad.createPatch(envs, sf)

		patchesBytes, err := json.Marshal(patches)
		if err != nil {
			fmt.Errorf("could not marshal JSON patch: %v", err)
		}

		return patchesBytes, nil
	}

	// 2.b
	return nil, fmt.Errorf("SNOWFLAKE_DATA_CENTER_ID and SNOWFLAKE_WORKER_ID should be set as a pair")
}

func removePod(ad *admission, pod *coreV1.Pod) {
	rsName := pod.OwnerReferences[0].Name
	envs := pod.Spec.Containers[0].Env

	var workerID int
	for _, env := range envs {
		if env.Name == "SNOWFLAKE_WORKER_ID" {
			workerID, _ = strconv.Atoi(env.Value)
			break
		}
	}

	var deletedPod bool

	mu.Lock()
	pa := populateApp(rsName)
	for i, n := range pa.Nodes {
		if pod.Spec.NodeName == n.Name {
			for j, p := range n.Pods {
				if workerID == p.ID {
					deletedPod = true
					l := len(n.Pods)
					n.Pods[j] = n.Pods[l-1]         // Copy last element to index i.
					pa.Nodes[i].Pods = n.Pods[:l-1] // Truncate slice.
					// pa.Nodes[i].Pods = append(n.Pods[:j], n.Pods[j+1:]...)
					log.Println("Deleted pod is {Node:", n.Name, n.ID, ", Pod:", n.Pods[j].ID, "}")
				}
			}
		}
		if deletedPod {
			break
		}
	}
	pa.saveApp()
	mu.Unlock()
}
