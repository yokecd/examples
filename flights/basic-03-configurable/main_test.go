package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplicaCount(t *testing.T) {
	deployment := CreateDeployment(DeploymentConfig{Replicas: 3})
	require.EqualValues(t, 3, *deployment.Spec.Replicas)
}
