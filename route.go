package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	admissionV1 "k8s.io/api/admission/v1"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func HandlePodValidate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	ad := admission{
		w: w,
	}

	ad.populateRequest(body)

	pod := &coreV1.Pod{}

	if err := json.Unmarshal(ad.review.Request.OldObject.Raw, pod); err != nil {
		fmt.Errorf("could not unmarshal pod on admission request: %v", err)
	}

	admissionReviewResponse := admissionV1.AdmissionReview{
		Response: &admissionV1.AdmissionResponse{
			UID:     ad.review.Request.UID,
			Allowed: true,
			// Warnings: warnings,
		},
	}

	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	if _, err := ad.w.Write(bytes); err != nil {
		fmt.Errorf("Error (http):", err)
	}
}

func HandlePodMutate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	ad := admission{
		w: w,
	}

	ad.populateRequest(body)

	// Default to permit admission webhook
	var allowed bool = true
	var status v1.Status
	var patchBytes []byte
	// warnings := []string{}

	patchBytes, err = HandlePod(&ad)

	if err != nil {
		allowed = false
		status = v1.Status{
			Message: err.Error(),
		}
	}

	ad.sendAdmissionResponse(allowed, patchBytes, &status)
}
