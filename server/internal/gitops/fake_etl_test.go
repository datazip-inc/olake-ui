package gitops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/datazip-inc/olake-ui/server/internal/constants"
	"github.com/datazip-inc/olake-ui/server/internal/models"
	"github.com/datazip-inc/olake-ui/server/internal/models/dto"
	"github.com/datazip-inc/olake-ui/server/internal/types"
	"github.com/datazip-inc/olake-ui/server/internal/utils"
)

// Operation ids the fake returns when a test connection or a stream difference starts. The
// reconciler then waits for the operation with AwaitOperationOutput, which answers from
// fakeETL.operationResults; a test overrides an entry to make that operation fail.
const (
	testSourceOperationID      = "op-test-source"
	testDestinationOperationID = "op-test-destination"
	streamDiffOperationID      = "op-stream-difference"
)

// fakeETL serves every service interface the reconcilers depend on.
var _ etlService = (*fakeETL)(nil)

// fakeETL is an in-memory ETL service. Unlike a stub, it keeps state: a create is visible to the
// next lookup, so a test can run several reconciles in a row (create, then edit, then retry) and
// see what an earlier one wrote, as the reconciler does against the real database.
//
// Every method follows the same shape: record logs the call and applies any injected failure, then
// the method reads or writes the stored rows.
type fakeETL struct {
	// The stored OLake entities, keyed by entity name, standing in for the database tables. A test
	// seeds them (addSource, addDestination, or a direct assignment) to start from an existing
	// entity, and reads them to check what a reconcile wrote.
	sources      map[string]*models.Source
	destinations map[string]*models.Destination
	jobs         map[string]*models.Job

	// lastID is the last entity id handed out. It starts at 100 so ids never match other small
	// numbers in the tests (user id 1, counts), and a wrong value fails an assertion.
	lastID int

	// failingMethods makes every call to a method fail: method name -> the error it returns. Use it
	// for an outage (database or Temporal down) or a rejection (invalid cron, duplicate name).
	failingMethods map[string]error

	// failCall fails one chosen call, where failingMethods fails every call to a method. It runs
	// at the start of every fake method and can count earlier calls with callCount, for example to
	// fail only the second GetJobByName, the read-back after a create.
	failCall func(method string) error

	// operationResults is what AwaitOperationOutput returns for each operation id: a test
	// connection's status output, the stream difference, or a workflow error. The defaults succeed.
	operationResults map[string]operationResult

	// discoveredStreams is the available streams every discover returns. Its value is
	// recognisable, so a test can tell a catalog that came from discover from one stored on the job.
	discoveredStreams string

	// calls logs every method call in order with its arguments. Read it with callCount, callArgs
	// (typed arguments) and writeCalls (only the calls that change OLake); testEnv.resetRecordings
	// clears it.
	calls []recordedCall
}

// recordedCall is one recorded method call.
type recordedCall struct {
	method string
	args   []any
}

// operationResult is the outcome of one operation: the workflow's output, or the error it failed with.
type operationResult struct {
	output map[string]interface{}
	err    error
}

func newFakeETL() *fakeETL {
	return &fakeETL{
		sources:        map[string]*models.Source{},
		destinations:   map[string]*models.Destination{},
		jobs:           map[string]*models.Job{},
		lastID:         100,
		failingMethods: map[string]error{},
		operationResults: map[string]operationResult{
			testSourceOperationID:      {output: connectionCheckOutput("SUCCEEDED", "")},
			testDestinationOperationID: {output: connectionCheckOutput("SUCCEEDED", "")},
			streamDiffOperationID:      {output: map[string]interface{}{"selected_streams": map[string]interface{}{"public": []interface{}{"users"}}}},
		},
		discoveredStreams: `{"streams":["discovered"]}`,
	}
}

// connectionCheckOutput is a check command's workflow output.
func connectionCheckOutput(status, message string) map[string]interface{} {
	return map[string]interface{}{"connectionStatus": map[string]interface{}{"status": status, "message": message}}
}

// record logs a call, then returns the error a test injected for it: failCall first, then
// failingMethods.
func (m *fakeETL) record(method string, args ...any) error {
	m.calls = append(m.calls, recordedCall{method: method, args: args})
	if m.failCall != nil {
		if err := m.failCall(method); err != nil {
			return err
		}
	}
	return m.failingMethods[method]
}

// callArgs returns the argument of type T from each call to method, in call order.
func callArgs[T any](m *fakeETL, method string) []T {
	var out []T
	for _, c := range m.calls {
		if c.method != method {
			continue
		}
		for _, arg := range c.args {
			if v, ok := arg.(T); ok {
				out = append(out, v)
			}
		}
	}
	return out
}

func (m *fakeETL) callCount(method string) int {
	n := 0
	for _, c := range m.calls {
		if c.method == method {
			n++
		}
	}
	return n
}

// writeCalls lists, in order, the calls that change OLake state.
func (m *fakeETL) writeCalls() []string {
	var out []string
	for _, c := range m.calls {
		switch c.method {
		case "CreateSource", "UpdateSource", "CreateDestination", "UpdateDestination", "CreateJob", "UpdateJob", "ActivateJob":
			out = append(out, c.method)
		}
	}
	return out
}

// newID hands out the next entity id.
func (m *fakeETL) newID() int {
	m.lastID++
	return m.lastID
}

// addSource seeds an existing postgres source, as if created earlier through the UI or a previous
// sync. Its config matches sourceConfigJSON, so a reconcile of sourceConfigJSON sees no change.
func (m *fakeETL) addSource(name, version string) *models.Source {
	src := &models.Source{
		ID: m.newID(), Name: name, ProjectID: testProjectID, Type: "postgres", Version: version,
		Config: `{"host":"pg","port":5432}`, BaseModel: models.BaseModel{UpdatedAt: time.Unix(1000, 0)},
	}
	m.sources[name] = src
	return src
}

// addDestination seeds an existing iceberg destination, the counterpart of addSource.
func (m *fakeETL) addDestination(name string) *models.Destination {
	dest := &models.Destination{
		ID: m.newID(), Name: name, ProjectID: testProjectID, DestType: "iceberg", Version: "v1.0.0",
		Config: `{"catalog_type":"glue"}`, BaseModel: models.BaseModel{UpdatedAt: time.Unix(1000, 0)},
	}
	m.destinations[name] = dest
	return dest
}

// findRow returns a copy of the first row that matches, as the database returns a new model per
// query, or notFound wrapped like the real service wraps it.
func findRow[T any](rows map[string]*T, match func(*T) bool, notFound error) (*T, error) {
	for _, row := range rows {
		if match(row) {
			cp := *row
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("%w: not in memory", notFound)
}

func (m *fakeETL) TestSourceConnection(_ context.Context, _ string, req *dto.SourceTestConnectionRequest) (string, error) {
	return testSourceOperationID, m.record("TestSourceConnection", req)
}

func (m *fakeETL) TestDestinationConnection(_ context.Context, _ string, req *dto.DestinationTestConnectionRequest) (string, error) {
	return testDestinationOperationID, m.record("TestDestinationConnection", req)
}

func (m *fakeETL) AwaitOperationOutput(_ context.Context, operationID string) (map[string]interface{}, error) {
	if err := m.record("AwaitOperationOutput", operationID); err != nil {
		return nil, err
	}
	result := m.operationResults[operationID]
	return result.output, result.err
}

// DiscoverWithCatalog keeps the given selection and returns m.discoveredStreams as the available streams,
// so a test can tell a fresh discover (empty stored catalog) from a merge.
func (m *fakeETL) DiscoverWithCatalog(_ context.Context, _ string, req *dto.StreamsRequest, stored types.StreamsCatalog) (types.StreamsCatalog, error) {
	if err := m.record("DiscoverWithCatalog", req, stored); err != nil {
		return types.StreamsCatalog{}, err
	}
	return types.StreamsCatalog{Available: m.discoveredStreams, Selected: stored.Selected}, nil
}

func (m *fakeETL) GetStreamDifference(_ context.Context, _ string, jobID int, req dto.StreamDifferenceRequest) (string, error) {
	return streamDiffOperationID, m.record("GetStreamDifference", jobID, req)
}

func (m *fakeETL) GetSourceByName(_ context.Context, projectID, name string) (*models.Source, error) {
	if err := m.record("GetSourceByName", name); err != nil {
		return nil, err
	}
	return findRow(m.sources, func(s *models.Source) bool { return s.Name == name && s.ProjectID == projectID }, constants.ErrSourceNotFound)
}

func (m *fakeETL) GetSourceByID(_ context.Context, projectID string, id int) (*models.Source, error) {
	if err := m.record("GetSourceByID", id); err != nil {
		return nil, err
	}
	return findRow(m.sources, func(s *models.Source) bool { return s.ID == id && s.ProjectID == projectID }, constants.ErrSourceNotFound)
}

func (m *fakeETL) CreateSource(_ context.Context, req *dto.CreateSourceRequest, projectID string, userID *int) error {
	if err := m.record("CreateSource", req, userID); err != nil {
		return err
	}
	m.sources[req.Name] = &models.Source{ID: m.newID(), Name: req.Name, ProjectID: projectID, Type: req.Type, Version: req.Version, Config: req.Config.String()}
	return nil
}

func (m *fakeETL) UpdateSource(_ context.Context, _ string, id int, req *dto.UpdateSourceRequest, userID *int) error {
	return m.record("UpdateSource", id, req, userID)
}

func (m *fakeETL) GetDestinationByName(_ context.Context, projectID, name string) (*models.Destination, error) {
	if err := m.record("GetDestinationByName", name); err != nil {
		return nil, err
	}
	return findRow(m.destinations, func(d *models.Destination) bool { return d.Name == name && d.ProjectID == projectID }, constants.ErrDestinationNotFound)
}

func (m *fakeETL) GetDestinationByID(_ context.Context, projectID string, id int) (*models.Destination, error) {
	if err := m.record("GetDestinationByID", id); err != nil {
		return nil, err
	}
	return findRow(m.destinations, func(d *models.Destination) bool { return d.ID == id && d.ProjectID == projectID }, constants.ErrDestinationNotFound)
}

func (m *fakeETL) CreateDestination(_ context.Context, req *dto.CreateDestinationRequest, projectID string, userID *int) error {
	if err := m.record("CreateDestination", req, userID); err != nil {
		return err
	}
	m.destinations[req.Name] = &models.Destination{ID: m.newID(), Name: req.Name, ProjectID: projectID, DestType: req.Type, Version: req.Version, Config: req.Config.String()}
	return nil
}

func (m *fakeETL) UpdateDestination(_ context.Context, id int, _ string, req *dto.UpdateDestinationRequest, userID *int) error {
	return m.record("UpdateDestination", id, req, userID)
}

func (m *fakeETL) GetJobByName(_ context.Context, projectID, name string) (*models.Job, error) {
	if err := m.record("GetJobByName", name); err != nil {
		return nil, err
	}
	return findRow(m.jobs, func(j *models.Job) bool { return j.Name == name && j.ProjectID == projectID }, constants.ErrJobNotFound)
}

func (m *fakeETL) CreateJob(_ context.Context, req *dto.CreateJobRequest, projectID string, userID *int) error {
	if err := m.record("CreateJob", req, userID); err != nil {
		return err
	}
	m.jobs[req.Name] = &models.Job{
		ID: m.newID(), Name: req.Name, ProjectID: projectID, SourceID: *req.Source.ID, DestID: *req.Destination.ID,
		Active: req.Activate, Frequency: req.Frequency, AdvancedSettings: advancedSettingsJSON(req.AdvancedSettings),
		AvailableStreamsConfig: utils.StringPtr(req.AvailableStreamsConfig), SelectedStreamsConfig: utils.StringPtr(req.SelectedStreamsConfig),
	}
	return nil
}

func (m *fakeETL) UpdateJob(_ context.Context, req *dto.UpdateJobRequest, _ string, jobID int, userID *int) error {
	if err := m.record("UpdateJob", jobID, req, userID); err != nil {
		return err
	}
	if job := m.jobByID(jobID); job != nil {
		job.Name, job.Frequency = req.Name, req.Frequency
		job.SourceID, job.DestID = *req.Source.ID, *req.Destination.ID
		job.AdvancedSettings = advancedSettingsJSON(req.AdvancedSettings)
		job.AvailableStreamsConfig = utils.StringPtr(req.AvailableStreamsConfig)
		job.SelectedStreamsConfig = utils.StringPtr(req.SelectedStreamsConfig)
	}
	return nil
}

func (m *fakeETL) ActivateJob(_ context.Context, jobID int, req dto.JobStatusRequest, userID *int) error {
	if err := m.record("ActivateJob", jobID, req, userID); err != nil {
		return err
	}
	if job := m.jobByID(jobID); job != nil {
		job.Active = req.Activate
	}
	return nil
}

func (m *fakeETL) jobByID(id int) *models.Job {
	for _, job := range m.jobs {
		if job.ID == id {
			return job
		}
	}
	return nil
}

// advancedSettingsJSON stores advanced settings as the job table does: JSON, or NULL when unset.
func advancedSettingsJSON(settings *dto.AdvancedSettings) *string {
	if settings == nil {
		return nil
	}
	b, _ := json.Marshal(settings)
	s := string(b)
	return &s
}
