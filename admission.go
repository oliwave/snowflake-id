package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	admissionV1 "k8s.io/api/admission/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type admission struct {
	review *admissionV1.AdmissionReview
	w      http.ResponseWriter
}

func (a *admission) populateRequest(admissionRequestBody []byte) {
	a.review = &admissionV1.AdmissionReview{}
	if _, _, err := universalDeserializer.Decode(admissionRequestBody, nil, a.review); err != nil {
		a.w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("could not deserialize request: %v", err)
	} else if a.review.Request == nil {
		a.w.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	// fmt.Printf("Type: %v \t Event: %v \t Name: %v \n",
	// 	a.review.Request.Kind,
	// 	a.review.Request.Operation,
	// 	a.review.Request.Name,
	// )
}

// func (a *admission) sendAdmissionResponse(allowed bool, patchBytes []byte, result *v1.Status, warnings []string) {
func (a *admission) sendAdmissionResponse(allowed bool, patchBytes []byte, result *v1.Status) {
	admissionReviewResponse := admissionV1.AdmissionReview{
		Response: &admissionV1.AdmissionResponse{
			UID:     a.review.Request.UID,
			Allowed: allowed,
			Patch:   patchBytes,
			Result:  result,
			// Warnings: warnings,
		},
	}

	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	if _, err := a.w.Write(bytes); err != nil {
		fmt.Errorf("Error (http):", err)
	}
}

func (s *admission) createPatch(envs []coreV1.EnvVar, sf *snowflake) []patchOperation {
	envs = append(envs, []coreV1.EnvVar{
		{
			Name:  "SNOWFLAKE_DATA_CENTER_ID",
			Value: strconv.Itoa(sf.datacenterId),
		},
		{
			Name:  "SNOWFLAKE_WORKER_ID",
			Value: strconv.Itoa(sf.workerId),
		},
	}...)

	return append([]patchOperation{}, []patchOperation{
		{
			Op:    "add",
			Path:  "/spec/containers/0/env",
			Value: envs,
		},
		{
			Op:    "add",
			Path:  "/spec/nodeName",
			Value: sf.nodeName,
		},
	}...)
}
