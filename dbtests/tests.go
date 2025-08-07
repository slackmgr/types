//nolint:goconst
package dbtests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	common "github.com/peteraglen/slack-manager-common"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAlert(t *testing.T, client common.DB) {
	ctx := context.Background()

	alert1 := newTestAlert("C0ABABABAB", ksuid.New().String())
	alert2 := newTestAlert("C0ABABABAB", ksuid.New().String())

	err := client.SaveAlert(ctx, alert1)
	require.NoError(t, err, "failed to save alert1")

	err = client.SaveAlert(ctx, alert2)
	require.NoError(t, err, "failed to save alert2")

	// Saving the same alert again should not fail
	err = client.SaveAlert(ctx, alert1)
	require.NoError(t, err, "failed to save alert1 again")

	err = client.SaveAlert(ctx, nil)
	require.Error(t, err, "should fail to save nil alert")
}

func TestSaveIssue(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	require := require.New(t)

	err := client.SaveIssue(ctx, nil)
	require.Error(err, "should fail to save nil issue")

	corr1 := ksuid.New().String()
	alert1 := newTestAlert(channel, corr1)
	issue1 := newTestIssue(alert1, ksuid.New().String())

	corr2 := ksuid.New().String()
	alert2 := newTestAlert(channel, corr2)
	issue2 := newTestIssue(alert2, ksuid.New().String())

	err = client.SaveIssue(ctx, issue1)
	require.NoError(err)
	id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel, corr1)
	require.NoError(err, "failed to get issue after saving")
	assert.Equal(t, issue1.ID, id, "issue ID should match after saving")
	require.NotNil(issueBody, "issue body should not be nil after saving")
	foundIssue := testIssueFromJSON(issueBody)
	assert.Equal(t, issue1.ID, foundIssue.ID, "issue ID should match after saving")
	assert.Equal(t, issue1.SlackPostID, foundIssue.SlackPostID, "SlackPostID should match after saving")

	err = client.SaveIssue(ctx, issue2)
	require.NoError(err)
	id, issueBody, err = client.FindOpenIssueByCorrelationID(ctx, channel, corr2)
	require.NoError(err, "failed to get issue after saving")
	assert.Equal(t, issue2.ID, id, "issue ID should match after saving")
	require.NotNil(issueBody, "issue body should not be nil after saving")
	foundIssue = testIssueFromJSON(issueBody)
	assert.Equal(t, issue2.ID, foundIssue.ID, "issue ID should match after saving")
	assert.Equal(t, issue2.SlackPostID, foundIssue.SlackPostID, "SlackPostID should match after saving")

	// Saving the same issue again should update the existing issue
	issue1.SlackPostID = ksuid.New().String() // Simulate a change in SlackPostID
	err = client.SaveIssue(ctx, issue1)
	require.NoError(err)
	id, issueBody, err = client.FindOpenIssueByCorrelationID(ctx, channel, corr1)
	require.NoError(err, "failed to get issue after saving")
	assert.Equal(t, issue1.ID, id, "issue ID should match after saving")
	require.NotNil(issueBody, "issue body should not be nil after saving")
	foundIssue = testIssueFromJSON(issueBody)
	assert.Equal(t, issue1.ID, foundIssue.ID, "issue ID should match after saving")
	assert.Equal(t, issue1.SlackPostID, foundIssue.SlackPostID, "SlackPostID should match after saving")
}

func TestMoveIssue(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel1 := "C0ABABABAB"
	channel2 := "C0ABABABAC"
	assert := assert.New(t)
	require := require.New(t)

	corr1 := ksuid.New().String()
	alert1 := newTestAlert(channel1, corr1)
	issue1 := newTestIssue(alert1, ksuid.New().String())

	corr2 := ksuid.New().String()
	alert2 := newTestAlert(channel1, corr2)
	issue2 := newTestIssue(alert2, ksuid.New().String())

	err := client.SaveIssues(ctx, issue1, issue2)
	require.NoError(err)
	issuesChannel1, err := client.LoadOpenIssuesInChannel(ctx, channel1)
	require.NoError(err, "should not error when loading open issues")
	assert.Len(issuesChannel1, 2, "should have 2 issues in channel1 after saving")
	issuesChannel2, err := client.LoadOpenIssuesInChannel(ctx, channel2)
	require.NoError(err, "should not error when loading open issues in channel2")
	assert.Empty(issuesChannel2, "should have 0 issues in channel2 after saving")

	// Simulate moving the issue to channel2
	issue1.LastAlert.SlackChannelID = channel2
	err = client.MoveIssue(ctx, issue1, channel1, channel2)
	require.NoError(err)

	// Verify that the issue was moved correctly
	issuesChannel1, err = client.LoadOpenIssuesInChannel(ctx, channel1)
	require.NoError(err, "should not error when loading open issues after moving")
	assert.Len(issuesChannel1, 1, "should have 1 issue left in channel1 after moving")
	issuesChannel2, err = client.LoadOpenIssuesInChannel(ctx, channel2)
	require.NoError(err, "should not error when loading open issues in channel2 after moving")
	assert.Len(issuesChannel2, 1, "should have 1 issue in channel2 after moving")

	// Verify that the moved issue cannot be found in the old channel
	id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel1, corr1)
	require.NoError(err, "failed to get issue after saving with updated channel ID")
	assert.Empty(id, "should not find issue in old channel after moving")
	assert.Nil(issueBody, "should not find issue body in old channel after moving")

	// Verify that the moved issue can be found in the new channel
	id, issueBody, err = client.FindOpenIssueByCorrelationID(ctx, channel2, corr1)
	require.NoError(err, "failed to get issue after moving to new channel")
	assert.NotEmpty(id, "should find issue in new channel after moving")
	assert.NotNil(issueBody, "should find issue body in new channel after moving")
}

func TestFindOpenIssueByCorrelationID(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	assert := assert.New(t)
	require := require.New(t)

	_, _, err := client.FindOpenIssueByCorrelationID(ctx, "", "foo")
	require.Error(err, "should fail to find issue with empty channel ID")

	_, _, err = client.FindOpenIssueByCorrelationID(ctx, channel, "")
	require.Error(err, "should fail to find issue with empty correlation ID")

	correlationID := ksuid.New().String()
	alert := newTestAlert(channel, correlationID)
	issue := newTestIssue(alert, ksuid.New().String())

	// Lookup by correlation ID before saving should return nil
	id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel, correlationID)
	require.NoError(err, "should not error when looking up issue by correlation ID before saving")
	assert.Empty(id, "should not have an ID before saving")
	assert.Nil(issueBody, "should not find issue by correlation ID before saving")

	// Save the issue
	err = client.SaveIssue(ctx, issue)
	require.NoError(err, "should not error when saving issue")

	// Lookup by correlation ID after saving should return the issue
	id, issueBody, err = client.FindOpenIssueByCorrelationID(ctx, channel, correlationID)
	require.NoError(err, "should not error when looking up issue by correlation ID after saving")
	assert.Equal(issue.ID, id, "should return the correct issue ID")
	require.NotNil(issueBody)
	foundIssue := testIssueFromJSON(issueBody)
	assert.Equal(issue.ID, foundIssue.ID)
	assert.Equal(issue.LastAlert.SlackChannelID, foundIssue.LastAlert.SlackChannelID)
	assert.Equal(issue.LastAlert.CorrelationID, foundIssue.LastAlert.CorrelationID)
	assert.Equal(issue.LastAlert.Header, foundIssue.LastAlert.Header)
	assert.Equal(issue.LastAlert.Text, foundIssue.LastAlert.Text)
	assert.Equal(issue.Archived, foundIssue.Archived)
	assert.Equal(issue.SlackPostID, foundIssue.SlackPostID)

	// An archived issue should not be found
	correlationIDArchived := ksuid.New().String()
	alertArchived := newTestAlert(channel, correlationIDArchived)
	issueArchived := newTestIssue(alertArchived, ksuid.New().String())
	issueArchived.Archived = true
	err = client.SaveIssue(ctx, issueArchived)
	require.NoError(err, "should not error when saving archived issue")
	id, issueBody, err = client.FindOpenIssueByCorrelationID(ctx, channel, correlationIDArchived)
	require.NoError(err, "should not error when looking up archived issue by correlation ID")
	assert.Empty(id, "should not return ID for archived issue")
	assert.Nil(issueBody, "should not find archived issue by correlation ID")
}

func TestFindIssueBySlackPostID(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	assert := assert.New(t)
	require := require.New(t)

	_, _, err := client.FindIssueBySlackPostID(ctx, "", "foo")
	require.Error(err, "should fail to find issue with empty channel ID")

	_, _, err = client.FindIssueBySlackPostID(ctx, channel, "")
	require.Error(err, "should fail to find issue with empty SlackPostID")

	alert := newTestAlert(channel, ksuid.New().String())
	postID := ksuid.New().String()
	issue := newTestIssue(alert, postID)

	// Lookup by SlackPostID before saving should return nil
	id, issueBody, err := client.FindIssueBySlackPostID(ctx, channel, postID)
	require.NoError(err, "should not error when looking up issue by SlackPostID before saving")
	assert.Empty(id, "should not have an ID before saving")
	assert.Nil(issueBody)

	// Save the issue
	err = client.SaveIssue(ctx, issue)
	require.NoError(err, "should not error when saving issue")

	// Lookup by SlackPostID after saving should return the issue
	id, issueBody, err = client.FindIssueBySlackPostID(ctx, channel, postID)
	require.NoError(err, "should not error when looking up issue by SlackPostID after saving")
	assert.Equal(issue.ID, id, "should return the correct issue ID")
	require.NotNil(issueBody)
	foundIssue := testIssueFromJSON(issueBody)
	assert.Equal(issue.ID, foundIssue.ID)
	assert.Equal(issue.LastAlert.SlackChannelID, foundIssue.LastAlert.SlackChannelID)
	assert.Equal(issue.LastAlert.CorrelationID, foundIssue.LastAlert.CorrelationID)
	assert.Equal(issue.LastAlert.Header, foundIssue.LastAlert.Header)
	assert.Equal(issue.LastAlert.Text, foundIssue.LastAlert.Text)
	assert.Equal(issue.Archived, foundIssue.Archived)
	assert.Equal(issue.SlackPostID, foundIssue.SlackPostID)

	// Saving the same issue again should update the existing issue
	newPostID := ksuid.New().String()
	issue.SlackPostID = newPostID // Simulate a change in SlackPostID
	err = client.SaveIssue(ctx, issue)
	require.NoError(err, "should not error when updating issue with new SlackPostID")

	// Lookup by old SlackPostID should return nil
	id, issueBody, err = client.FindIssueBySlackPostID(ctx, channel, postID)
	require.NoError(err, "should not error when looking up issue by old SlackPostID")
	assert.Empty(id, "should not return ID for old SlackPostID")
	assert.Nil(issueBody)

	// Lookup by new SlackPostID should return the updated issue
	id, issueBody, err = client.FindIssueBySlackPostID(ctx, channel, newPostID)
	require.NoError(err, "should not error when looking up issue by updated SlackPostID")
	assert.Equal(issue.ID, id, "should return the correct issue ID after update")
	assert.NotNil(issueBody)
	foundIssue = testIssueFromJSON(issueBody)
	assert.Equal(issue.ID, foundIssue.ID)
	assert.Equal(issue.SlackPostID, foundIssue.SlackPostID)
}

func TestSaveIssues(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	assert := assert.New(t)
	require := require.New(t)

	// Ensure the db is empty before starting, as this test assumes a clean state
	err := client.DropAllData(ctx)
	require.NoError(err, "should not error when dropping all tables before test")

	err = client.Init(ctx, true)
	require.NoError(err, "should not error when initializing client")

	err = client.SaveIssues(ctx)
	require.NoError(err, "should not error when updating with empty issues list")

	issue1 := newTestIssue(newTestAlert(channel, ksuid.New().String()), ksuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel, ksuid.New().String()), ksuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel, ksuid.New().String()), ksuid.New().String())

	// Save the issues
	err = client.SaveIssues(ctx, issue1, issue2, issue3)
	require.NoError(err, "should not error when saving multiple issues")

	// Verify that the issues were saved correctly
	issues, err := client.LoadOpenIssuesInChannel(ctx, channel)
	require.NoError(err, "should not error when loading open issues")
	assert.Len(issues, 3, "should have 3 issues after saving")
}

func TestFindActiveChannels(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel1 := "C0ABABABAB"
	channel2 := "C0ABABABAC"
	channel3 := "C0ABABABAD"
	assert := assert.New(t)
	require := require.New(t)

	// Ensure the db is empty before starting, as this test assumes a clean state
	err := client.DropAllData(ctx)
	require.NoError(err, "should not error when dropping all tables before test")

	err = client.Init(ctx, true)
	require.NoError(err, "should not error when initializing client")

	issue1 := newTestIssue(newTestAlert(channel1, ksuid.New().String()), ksuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel2, ksuid.New().String()), ksuid.New().String())
	issue2a := newTestIssue(newTestAlert(channel2, ksuid.New().String()), ksuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel3, ksuid.New().String()), ksuid.New().String())
	issue3a := newTestIssue(newTestAlert(channel3, ksuid.New().String()), ksuid.New().String())
	issue3.Archived = true // Mark one issue as archived

	// Save the issues
	err = client.SaveIssues(ctx, issue1, issue2, issue2a, issue3, issue3a)
	require.NoError(err, "should not error when saving multiple issues")

	// Verify that active channels are found correctly
	issues, err := client.FindActiveChannels(ctx)
	require.NoError(err, "should not error when finding active channels")
	assert.Len(issues, 3, "should have 3 active channels after saving issues")

	issue1.Archived = true  // Mark issue1 as archived
	issue2.Archived = true  // Mark issue2 as archived
	issue2a.Archived = true // Mark issue2a as archived

	err = client.SaveIssues(ctx, issue1, issue2, issue2a)
	require.NoError(err, "should not error when updating issues to archived")

	// Verify that no open issues are loaded after archiving all
	issues, err = client.FindActiveChannels(ctx)
	require.NoError(err, "should not error when finding active channels after archiving")
	assert.Len(issues, 1, "should have 1 active channel after archiving all issues in channel1 and channel2")
}

func TestLoadOpenIssuesInChannel(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel1 := "C0ABABABAB"
	channel2 := "C0ABABABAC"
	assert := assert.New(t)
	require := require.New(t)

	// Ensure the db is empty before starting, as this test assumes a clean state
	err := client.DropAllData(ctx)
	require.NoError(err, "should not error when dropping all tables before test")

	err = client.Init(ctx, true)
	require.NoError(err, "should not error when initializing client")

	issue1 := newTestIssue(newTestAlert(channel1, ksuid.New().String()), ksuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel1, ksuid.New().String()), ksuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel2, ksuid.New().String()), ksuid.New().String())
	issue4 := newTestIssue(newTestAlert(channel2, ksuid.New().String()), ksuid.New().String())
	issue4.Archived = true // Mark one issue as archived

	// Save the issues for channel1
	err = client.SaveIssues(ctx, issue1, issue2, issue3, issue4)
	require.NoError(err, "should not error when saving multiple issues")

	// Verify that only open issues are loaded for channel1
	issues, err := client.LoadOpenIssuesInChannel(ctx, channel1)
	require.NoError(err, "should not error when loading open issues")
	assert.Len(issues, 2, "should have 2 open issues after saving for channel1")
	issuesChannel1 := testIssuesFromJSON(issues)
	assert.Contains(issuesChannel1, issue1.UniqueID(), "should contain issue1 in channel1")
	assert.Contains(issuesChannel1, issue2.UniqueID(), "should contain issue2 in channel1")

	// Verify that only open issues are loaded for channel2
	issues, err = client.LoadOpenIssuesInChannel(ctx, channel2)
	require.NoError(err, "should not error when loading open issues in channel2")
	assert.Len(issues, 1, "should have 1 open issue in channel2 after saving")
	issuesChannel2 := testIssuesFromJSON(issues)
	assert.Contains(issuesChannel2, issue3.UniqueID(), "should contain issue3 in channel2")

	// Verify that no open issues are loaded for channel2 after archiving all
	issue3.Archived = true
	err = client.SaveIssues(ctx, issue3)
	require.NoError(err, "should not error when updating issue3 to archived")
	issues, err = client.LoadOpenIssuesInChannel(ctx, channel2)
	require.NoError(err, "should not error when loading open issues in channel2 after archiving")
	assert.Empty(issues, "should have 0 open issues in channel2 after archiving all")
}

func TestCreatingAndFindingMoveMappings(t *testing.T, client common.DB) {
	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)

	correlationID := ksuid.New().String()
	originalChannelID := "C0ABABABAB"
	targetChannelID := "C0ABABABAC"

	moveMapping := newTestMoveMapping(correlationID, originalChannelID, targetChannelID)
	err := client.SaveMoveMapping(ctx, moveMapping)
	require.NoError(err, "should not error when creating move mapping")

	// Verify that the move mapping was saved correctly
	moveMappingBody, err := client.FindMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when finding move mapping")
	assert.NotNil(moveMappingBody, "should find move mapping after saving")
	foundMoveMapping := moveMappingFromJSON(moveMappingBody)
	assert.Equal(moveMapping.ID, foundMoveMapping.ID, "move mapping ID should match")
	assert.Equal(moveMapping.OriginalChannelID, foundMoveMapping.OriginalChannelID, "original channel ID should match")
	assert.Equal(moveMapping.TargetChannelID, foundMoveMapping.TargetChannelID, "target channel ID should match")
	assert.Equal(moveMapping.CorrelationID, foundMoveMapping.CorrelationID, "correlation ID should match")
	assert.Equal(moveMapping.Timestamp.UTC().Format(time.RFC3339Nano), foundMoveMapping.Timestamp.UTC().Format(time.RFC3339Nano), "timestamp should match")

	moveMapping.TargetChannelID = "C0ABABABAD" // Simulate a change in target channel ID
	err = client.SaveMoveMapping(ctx, moveMapping)
	require.NoError(err, "should not error when updating existing move mapping")

	// Verify that the move mapping was updated correctly
	moveMappingBody, err = client.FindMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when finding updated move mapping")
	assert.NotNil(moveMappingBody, "should find updated move mapping after saving")
	foundMoveMapping = moveMappingFromJSON(moveMappingBody)
	assert.Equal(moveMapping.ID, foundMoveMapping.ID, "move mapping ID should still match after update")
	assert.Equal(moveMapping.OriginalChannelID, foundMoveMapping.OriginalChannelID, "original channel ID should still match after update")
	assert.Equal(moveMapping.TargetChannelID, foundMoveMapping.TargetChannelID, "target channel ID should match after update")
	assert.Equal(moveMapping.CorrelationID, foundMoveMapping.CorrelationID, "correlation ID should still match after update")
	assert.Equal(moveMapping.Timestamp.UTC().Format(time.RFC3339Nano), foundMoveMapping.Timestamp.UTC().Format(time.RFC3339Nano), "timestamp should still match after update")

	err = client.SaveMoveMapping(ctx, nil)
	require.Error(err, "should fail to create move mapping with nil move mapping")

	_, err = client.FindMoveMapping(ctx, "", correlationID)
	require.Error(err, "should fail to find move mapping with empty original channel ID")

	_, err = client.FindMoveMapping(ctx, originalChannelID, "")
	require.Error(err, "should fail to find move mapping with empty correlation ID")

	moveMappingBody, err = client.FindMoveMapping(ctx, "foo", correlationID)
	require.NoError(err, "should not error when finding move mapping with invalid original channel ID")
	assert.Nil(moveMappingBody, "should not find move mapping with invalid original channel ID")

	moveMappingBody, err = client.FindMoveMapping(ctx, originalChannelID, "foo")
	require.NoError(err, "should not error when finding move mapping with invalid correlation ID")
	assert.Nil(moveMappingBody, "should not find move mapping with invalid correlation ID")
}

func TestCreatingAndFindingChannelProcessingState(t *testing.T, client common.DB) {
	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)
	channelID := "C0ABABABAB"
	now := time.Now()

	// Create a new channel processing state
	state := common.NewChannelProcessingState(channelID)
	state.LastProcessed = now

	err := client.SaveChannelProcessingState(ctx, state)
	require.NoError(err, "should not error when saving channel processing state")

	// Verify that the channel processing state was saved correctly
	foundState, err := client.FindChannelProcessingState(ctx, channelID)
	require.NoError(err, "should not error when finding channel processing state")
	assert.NotNil(foundState, "should find channel processing state after saving")
	assert.Equal(state.ChannelID, foundState.ChannelID, "channel ID should match")
	assert.Equal(state.Created.UTC().Format(time.RFC3339Nano), foundState.Created.UTC().Format(time.RFC3339Nano), "created timestamp should match")
	assert.Equal(state.LastProcessed.UTC().Format(time.RFC3339Nano), foundState.LastProcessed.UTC().Format(time.RFC3339Nano), "last processed timestamp should match")

	// Update the channel processing state
	state.LastProcessed = now.Add(5 * time.Minute)

	err = client.SaveChannelProcessingState(ctx, state)
	require.NoError(err, "should not error when updating channel processing state")
	foundState, err = client.FindChannelProcessingState(ctx, channelID)
	require.NoError(err, "should not error when finding updated channel processing state")
	assert.NotNil(foundState, "should find updated channel processing state after saving")
	assert.Equal(state.ChannelID, foundState.ChannelID, "channel ID should still match after update")
	assert.Equal(state.Created.UTC().Format(time.RFC3339Nano), foundState.Created.UTC().Format(time.RFC3339Nano), "created timestamp should still match after update")
	assert.Equal(state.LastProcessed.UTC().Format(time.RFC3339Nano), foundState.LastProcessed.UTC().Format(time.RFC3339Nano), "last processed timestamp should match after update")
}

type testIssue struct {
	ID            string        `json:"id"`
	CorrelationID string        `json:"correlationId"`
	LastAlert     *common.Alert `json:"lastAlert"`
	Archived      bool          `json:"archived"`
	SlackPostID   string        `json:"slackPostId"`
}

func newTestAlert(channelID, correlationID string) *common.Alert {
	alert := common.NewErrorAlert()
	alert.SlackChannelID = channelID
	alert.CorrelationID = correlationID
	alert.Header = "Test Alert"
	alert.Text = "This is a test alert"
	return alert
}

func newTestIssue(alert *common.Alert, slackPostID string) *testIssue {
	return &testIssue{
		ID:            alert.CorrelationID + alert.SlackChannelID + time.Now().UTC().Format(time.RFC3339Nano),
		CorrelationID: alert.CorrelationID,
		LastAlert:     alert,
		Archived:      false,
		SlackPostID:   slackPostID,
	}
}

func testIssueFromJSON(data []byte) *testIssue {
	var issue testIssue
	err := json.Unmarshal(data, &issue)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal issue: %v", err))
	}
	return &issue
}

func testIssuesFromJSON(issueBodies map[string]json.RawMessage) map[string]*testIssue {
	issues := make(map[string]*testIssue)
	for id, body := range issueBodies {
		issue := testIssueFromJSON(body)
		if issue != nil {
			issues[id] = issue
		}
	}
	return issues
}

func (issue *testIssue) ChannelID() string {
	return issue.LastAlert.SlackChannelID
}

func (issue *testIssue) UniqueID() string {
	return issue.ID
}

func (issue *testIssue) GetCorrelationID() string {
	return issue.CorrelationID
}

func (issue *testIssue) IsOpen() bool {
	return !issue.Archived
}

func (issue *testIssue) CurrentPostID() string {
	return issue.SlackPostID
}

func (issue *testIssue) MarshalJSON() ([]byte, error) {
	type Alias testIssue

	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(issue),
	})
}

type testMoveMapping struct {
	ID                string    `json:"id"`
	Timestamp         time.Time `json:"timestamp"`
	CorrelationID     string    `json:"correlationId"`
	OriginalChannelID string    `json:"originalChannelId"`
	TargetChannelID   string    `json:"targetChannelId"`
}

func newTestMoveMapping(correlationID, originalChannelID, targetChannelID string) *testMoveMapping {
	return &testMoveMapping{
		ID:                originalChannelID + "-" + correlationID,
		Timestamp:         time.Now(),
		CorrelationID:     correlationID,
		OriginalChannelID: originalChannelID,
		TargetChannelID:   targetChannelID,
	}
}

func (m *testMoveMapping) ChannelID() string {
	return m.OriginalChannelID
}

func (m *testMoveMapping) UniqueID() string {
	return m.ID
}

func (m *testMoveMapping) GetCorrelationID() string {
	return m.CorrelationID
}

func (m *testMoveMapping) MarshalJSON() ([]byte, error) {
	type Alias testMoveMapping

	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

func moveMappingFromJSON(data []byte) *testMoveMapping {
	var moveMapping testMoveMapping
	err := json.Unmarshal(data, &moveMapping)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal move mapping: %v", err))
	}
	return &moveMapping
}
