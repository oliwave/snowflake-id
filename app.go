package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	dapr "github.com/dapr/go-sdk/client"
)

const (
	cachedState = `cachedstate`
)

var (
	daprClient dapr.Client
	daprPort   = "3500"
)

type pod struct {
	ID int `json:"id"`
}

type node struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Pods pods   `json:"pods"`
}

type App struct {
	ReplicaSetName string `json:"replicaSetName"`
	Nodes          nodes  `json:"nodes"`
}

type pods []pod

type nodes []node

func (p pods) Len() int           { return len(p) }
func (p pods) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p pods) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (n nodes) Len() int           { return len(n) }
func (n nodes) Less(i, j int) bool { return n[i].ID < n[j].ID }
func (n nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

func populateApp(replicaSetName string) *App {
	ctx := context.Background()

	a := &App{
		ReplicaSetName: replicaSetName,
	}

	result, err := daprClient.GetState(ctx, cachedState, a.ReplicaSetName)
	if err != nil {
		panic(err)
	}

	if len(result.Value) != 0 {
		if err := json.Unmarshal(result.Value, a); err != nil {
			log.Fatalf("Error happened in JSON unmarshal. Err: %s", err)
		}
	}

	return a
}

func (a *App) saveApp() bool {
	ctx := context.Background()

	jsonData, err := json.Marshal(a)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	if err := daprClient.SaveState(ctx, cachedState, a.ReplicaSetName, jsonData); err != nil {
		panic(err)
	}

	return true
}

func startDapr() {
	client, err := dapr.NewClientWithPort(daprPort)
	if err != nil {
		panic(err)
	}

	daprClient = client
}

func init() {
	if port := os.Getenv("DAPR_GRPC_PORT"); len(port) != 0 {
		daprPort = port
	}
	fmt.Println("dapr port:", daprPort)
	startDapr()
}
