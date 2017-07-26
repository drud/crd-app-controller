package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// App represents app object
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AppSpec   `json:"spec"`
	Status            AppStatus `json:"status,omitempty"`
}

// AppSpec represents the fields for App
type AppSpec struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// AppStatus represents status info for App
type AppStatus struct {
	State   AppState `json:"state,omitempty"`
	Message string   `json:"message,omitempty"`
}

// AppState represents applications' state
type AppState string

const (
	// AppStateCreated is state when app is first created
	AppStateCreated AppState = "Created"
	// AppStateInvalidType is state when an invalid app type is defined
	AppStateInvalidType AppState = "Invalid App Type"
	// AppStateFailed is state when deploy failed
	AppStateFailed AppState = "Failed"
	// AppStateProcessed is state when app has been deployed
	AppStateProcessed AppState = "Processed"
)

// AppList represents a list of applications
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []App `json:"items"`
}
