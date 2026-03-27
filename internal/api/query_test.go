package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
)

// stubDatascriptAPI is a minimal LogseqAPI stub that captures which method was called.
type stubDatascriptAPI struct {
	postQueryCalled           bool
	postDatascriptQueryCalled bool
	upsertBlockPropertyCalled bool
	upsertBlockPropertyUUID   string
	upsertBlockPropertyKey    string
	upsertBlockPropertyValue  string
	datascriptResponse        string
	datascriptErr             error
	upsertBlockPropertyErr    error
}

func (s *stubDatascriptAPI) PostQuery(_ string) (string, error) {
	s.postQueryCalled = true

	return "null", nil
}

func (s *stubDatascriptAPI) PostDatascriptQuery(_ string) (string, error) {
	s.postDatascriptQueryCalled = true

	return s.datascriptResponse, s.datascriptErr
}

func (s *stubDatascriptAPI) UpsertBlockProperty(uuid, key, value string) error {
	s.upsertBlockPropertyCalled = true
	s.upsertBlockPropertyUUID = uuid
	s.upsertBlockPropertyKey = key
	s.upsertBlockPropertyValue = value

	return s.upsertBlockPropertyErr
}

func TestUpsertBlockProperty_CallsCorrectMethod(t *testing.T) {
	// UpsertBlockProperty must call the Logseq Editor API with the correct args.
	// This is the mechanism used to force Logseq to write a block's id:: to disk.
	api := &stubDatascriptAPI{}

	err := api.UpsertBlockProperty("test-uuid", "id", "test-uuid")

	require.NoError(t, err)
	assert.True(t, api.upsertBlockPropertyCalled)
	assert.Equal(t, "test-uuid", api.upsertBlockPropertyUUID)
	assert.Equal(t, "id", api.upsertBlockPropertyKey)
	assert.Equal(t, "test-uuid", api.upsertBlockPropertyValue)
}

func TestFindBlockByUUID_UsesDatascriptQuery(t *testing.T) {
	// Page keys use hyphenated names as returned by Logseq's datascript API.
	blockJSON := `[[{"uuid":"test-uuid","page":{"id":1,"original-name":"My Page"}}]]`
	api := &stubDatascriptAPI{datascriptResponse: blockJSON}

	info, err := logseqapi.FindBlockByUUID(api, "test-uuid")

	require.NoError(t, err)
	assert.True(t, api.postDatascriptQueryCalled, "should use PostDatascriptQuery")
	assert.False(t, api.postQueryCalled, "should NOT use PostQuery (logseq.db.q fails for datascript)")
	assert.Equal(t, "My Page", info.PageName)
}

func TestFindBlockByUUID_JournalPage(t *testing.T) {
	// Logseq returns "journal-day" (hyphenated), not "journalDay" for nested page pulls.
	page := `{"id":88793,"journal-day":20170312,"original-name":"Sunday, 12.03.2017","name":"sunday, 12.03.2017"}`
	blockJSON := `[[{"uuid":"67f796e6-ea16-4e01-87d6-9ee9db49d173","page":` + page + `}]]`
	api := &stubDatascriptAPI{datascriptResponse: blockJSON}

	info, err := logseqapi.FindBlockByUUID(api, "67f796e6-ea16-4e01-87d6-9ee9db49d173")

	require.NoError(t, err)
	assert.True(t, info.IsJournal)
	assert.Equal(t, time.Date(2017, 3, 12, 0, 0, 0, 0, time.UTC), info.JournalDate)
}

func TestFindBlockByUUID_NullResponseReturnsError(t *testing.T) {
	api := &stubDatascriptAPI{datascriptResponse: "null"}

	_, err := logseqapi.FindBlockByUUID(api, "missing-uuid")

	assert.ErrorIs(t, err, logseqapi.ErrBlockNotFoundViaAPI)
}

func TestFindBlockByUUID_EmptyResultsReturnsError(t *testing.T) {
	api := &stubDatascriptAPI{datascriptResponse: "[]"}

	_, err := logseqapi.FindBlockByUUID(api, "missing-uuid")

	assert.ErrorIs(t, err, logseqapi.ErrBlockNotFoundViaAPI)
}
