package gitops

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
)

// The create, skip and failure rules are shared with sources (reconcileResource) and tested in
// source_reconciler_test.go; these tests cover the destination wiring.
func TestDestinationReconcile(t *testing.T) {
	const spec = `{"name":"lake","type":"iceberg","version":"v1.0.0","config":{"catalog_type":"glue"}}`

	t.Run("create", func(t *testing.T) {
		e := newTestEnv(t, managedConfigMap(KindDestination, "lake-cm", manifestData(spec)))
		e.reconcileDestination("lake-cm")
		assert.Equal(t, []*dto.CreateDestinationRequest{{Name: "lake", Type: "iceberg", Version: "v1.0.0", Config: `{"catalog_type":"glue"}`}},
			callArgs[*dto.CreateDestinationRequest](e.etl, "CreateDestination"))
		// a standalone destination test has no source to check against, so SourceType is empty
		assert.Equal(t, []*dto.DestinationTestConnectionRequest{{Type: "iceberg", Version: "v1.0.0", Config: `{"catalog_type":"glue"}`}},
			callArgs[*dto.DestinationTestConnectionRequest](e.etl, "TestDestinationConnection"))
		assert.Equal(t, idString(e.etl.destinations["lake"].ID), e.annotationsOf("lake-cm")[AnnotationEntityID])
	})

	t.Run("changed config is updated", func(t *testing.T) {
		e := newTestEnv(t, managedConfigMap(KindDestination, "lake-cm", manifestData(`{"name":"lake","type":"iceberg","version":"v1.0.0","config":{"catalog_type":"rest"}}`)))
		dest := e.etl.addDestination("lake")
		e.reconcileDestination("lake-cm")
		assert.Equal(t, []int{dest.ID}, callArgs[int](e.etl, "UpdateDestination"))
		assert.Equal(t, []*dto.UpdateDestinationRequest{{Name: "lake", Type: "iceberg", Version: "v1.0.0", Config: `{"catalog_type":"rest"}`}},
			callArgs[*dto.UpdateDestinationRequest](e.etl, "UpdateDestination"))
	})

	t.Run("type compares against the stored dest_type", func(t *testing.T) {
		e := newTestEnv(t, managedConfigMap(KindDestination, "lake-cm", manifestData(`{"name":"lake","type":"iceberg","version":"v1.0.0","config":{"catalog_type":"glue"}}`)))
		e.etl.addDestination("lake")
		e.reconcileDestination("lake-cm")
		assert.Empty(t, e.etl.writeCalls())
		assert.Zero(t, e.etl.callCount("TestDestinationConnection"))
	})
}
