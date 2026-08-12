// Package v1 contains OLake GitOps API types used by the embedded reconciler.
// +groupName=olake.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Group   = "olake.io"
	Version = "v1"
)

var (
	GroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)
