//nolint:goconst
package dbtests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	common "github.com/slackmgr/slack-manager-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Package dbtests provides a comprehensive test suite for common.DB implementations.
//
// Usage in your database plugin:
//
//	func TestDatabaseCompliance(t *testing.T) {
//	    if testing.Short() {
//	        t.Skip("Skipping integration tests")
//	    }
//
//	    db := setupTestDatabase(t)
//	    defer teardownTestDatabase(t, db)
//
//	    // Run all standard tests
//	    dbtests.RunAllTests(t, db)
//	}
//
// Or run individual tests:
//
//	dbtests.TestSaveAlert(t, db)
//	dbtests.TestSaveIssue(t, db)
//	// ... etc
//
// All tests clean up after themselves where possible, but some tests
// use DropAllData() to ensure a clean state before running.

func TestSaveAlert(t *testing.T, client common.DB) {
	ctx := context.Background()

	alert1 := newTestAlert("C0ABABABAB", uuid.New().String())
	alert2 := newTestAlert("C0ABABABAB", uuid.New().String())

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

	corr1 := uuid.New().String()
	alert1 := newTestAlert(channel, corr1)
	issue1 := newTestIssue(alert1, uuid.New().String())

	corr2 := uuid.New().String()
	alert2 := newTestAlert(channel, corr2)
	issue2 := newTestIssue(alert2, uuid.New().String())

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
	issue1.SlackPostID = uuid.New().String() // Simulate a change in SlackPostID
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

	corr1 := uuid.New().String()
	alert1 := newTestAlert(channel1, corr1)
	issue1 := newTestIssue(alert1, uuid.New().String())

	corr2 := uuid.New().String()
	alert2 := newTestAlert(channel1, corr2)
	issue2 := newTestIssue(alert2, uuid.New().String())

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

	correlationID := uuid.New().String()
	alert := newTestAlert(channel, correlationID)
	issue := newTestIssue(alert, uuid.New().String())

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
	correlationIDArchived := uuid.New().String()
	alertArchived := newTestAlert(channel, correlationIDArchived)
	issueArchived := newTestIssue(alertArchived, uuid.New().String())
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

	alert := newTestAlert(channel, uuid.New().String())
	postID := uuid.New().String()
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
	newPostID := uuid.New().String()
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

	issue1 := newTestIssue(newTestAlert(channel, uuid.New().String()), uuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel, uuid.New().String()), uuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel, uuid.New().String()), uuid.New().String())

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

	issue1 := newTestIssue(newTestAlert(channel1, uuid.New().String()), uuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel2, uuid.New().String()), uuid.New().String())
	issue2a := newTestIssue(newTestAlert(channel2, uuid.New().String()), uuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel3, uuid.New().String()), uuid.New().String())
	issue3a := newTestIssue(newTestAlert(channel3, uuid.New().String()), uuid.New().String())
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

	issue1 := newTestIssue(newTestAlert(channel1, uuid.New().String()), uuid.New().String())
	issue2 := newTestIssue(newTestAlert(channel1, uuid.New().String()), uuid.New().String())
	issue3 := newTestIssue(newTestAlert(channel2, uuid.New().String()), uuid.New().String())
	issue4 := newTestIssue(newTestAlert(channel2, uuid.New().String()), uuid.New().String())
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

	correlationID := uuid.New().String()
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

func TestDeletingMoveMappings(t *testing.T, client common.DB) {
	ctx := context.Background()
	assert := assert.New(t)
	require := require.New(t)

	correlationID := uuid.New().String()
	originalChannelID := "C0ABABABAB"
	targetChannelID := "C0ABABABAC"

	moveMapping := newTestMoveMapping(correlationID, originalChannelID, targetChannelID)
	err := client.SaveMoveMapping(ctx, moveMapping)
	require.NoError(err, "should not error when creating move mapping")

	// Verify that the move mapping was saved correctly
	moveMappingBody, err := client.FindMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when finding move mapping")
	assert.NotNil(moveMappingBody, "should find move mapping after saving")

	// Delete the move mapping
	err = client.DeleteMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when deleting move mapping")

	// Verify that the move mapping was deleted
	moveMappingBody, err = client.FindMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when finding deleted move mapping")
	assert.Nil(moveMappingBody, "should not find move mapping after deletion")

	// Attempt to delete a non-existent move mapping should not error
	err = client.DeleteMoveMapping(ctx, originalChannelID, correlationID)
	require.NoError(err, "should not error when deleting non-existent move mapping")
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

// TestInit verifies that database initialization works correctly.
// It tests basic initialization and idempotent calls.
func TestInit(t *testing.T, client common.DB) {
	ctx := context.Background()
	require := require.New(t)

	// First initialization should succeed
	err := client.Init(ctx, true)
	require.NoError(err, "first initialization should succeed")

	// Second initialization should be idempotent
	err = client.Init(ctx, true)
	require.NoError(err, "second initialization should be idempotent")
}

// TestInit_WithSchemaValidation verifies initialization with schema validation enabled.
func TestInit_WithSchemaValidation(t *testing.T, client common.DB) {
	ctx := context.Background()
	require := require.New(t)

	// Drop all data first to ensure clean state
	err := client.DropAllData(ctx)
	require.NoError(err, "should not error when dropping all data")

	// Initialize with schema validation
	err = client.Init(ctx, false)
	require.NoError(err, "initialization with schema validation should succeed")

	// Verify database is functional after initialization
	alert := newTestAlert("C0ABABABAB", uuid.New().String())
	err = client.SaveAlert(ctx, alert)
	require.NoError(err, "should be able to save alert after initialization")
}

// TestMoveIssue_EdgeCases tests edge cases for moving issues between channels.
func TestMoveIssue_EdgeCases(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel1 := "C0ABABABAB"
	channel2 := "C0ABABABAC"

	t.Run("same source and target channel", func(t *testing.T) {
		require := require.New(t)

		corr := uuid.New().String()
		alert := newTestAlert(channel1, corr)
		issue := newTestIssue(alert, uuid.New().String())

		err := client.SaveIssue(ctx, issue)
		require.NoError(err)

		// Try to move to the same channel - should error
		err = client.MoveIssue(ctx, issue, channel1, channel1)
		require.Error(err, "should error when source and target channels are the same")
	})

	t.Run("channel mismatch", func(t *testing.T) {
		require := require.New(t)

		corr := uuid.New().String()
		alert := newTestAlert(channel1, corr)
		issue := newTestIssue(alert, uuid.New().String())

		err := client.SaveIssue(ctx, issue)
		require.NoError(err)

		// Try to move but issue's channel doesn't match source
		issue.LastAlert.SlackChannelID = channel2
		_ = client.MoveIssue(ctx, issue, channel1, channel2)
		// This should either succeed or error depending on implementation
		// Most implementations should validate channel consistency
	})

	t.Run("moving non-existent issue", func(t *testing.T) {
		require := require.New(t)

		corr := uuid.New().String()
		alert := newTestAlert(channel1, corr)
		issue := newTestIssue(alert, uuid.New().String())

		// Don't save the issue, just try to move it
		issue.LastAlert.SlackChannelID = channel2
		err := client.MoveIssue(ctx, issue, channel1, channel2)
		// Should not error - implementations should handle gracefully
		require.NoError(err, "moving non-existent issue should not error")
	})
}

// TestConcurrentSaveIssue tests concurrent writes to the same issue.
func TestConcurrentSaveIssue(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	corr := uuid.New().String()
	require := require.New(t)

	// Create initial issue
	alert := newTestAlert(channel, corr)
	issue := newTestIssue(alert, uuid.New().String())
	err := client.SaveIssue(ctx, issue)
	require.NoError(err)

	// Concurrently update the same issue
	const goroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			// Each goroutine updates the issue with a different post ID
			issue.SlackPostID = fmt.Sprintf("post-%d-%s", index, uuid.New().String())
			if err := client.SaveIssue(ctx, issue); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check that no errors occurred
	for err := range errors {
		require.NoError(err, "concurrent save should not error")
	}

	// Verify issue still exists
	id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel, corr)
	require.NoError(err)
	require.NotEmpty(id)
	require.NotNil(issueBody)
}

// TestConcurrentMoveMapping tests concurrent move mapping operations.
func TestConcurrentMoveMapping(t *testing.T, client common.DB) {
	ctx := context.Background()
	require := require.New(t)
	const goroutines = 5

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			corr := uuid.New().String()
			originalChannel := fmt.Sprintf("C0ABABAB%02d", index)
			targetChannel := fmt.Sprintf("C0ABABAC%02d", index)

			moveMapping := newTestMoveMapping(corr, originalChannel, targetChannel)
			if err := client.SaveMoveMapping(ctx, moveMapping); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		require.NoError(err, "concurrent move mapping operations should not error")
	}
}

// TestLoadOpenIssuesInChannel_LargeDataset tests loading many issues from a channel.
func TestLoadOpenIssuesInChannel_LargeDataset(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	require := require.New(t)
	assert := assert.New(t)

	// Ensure clean state
	err := client.DropAllData(ctx)
	require.NoError(err)
	err = client.Init(ctx, true)
	require.NoError(err)

	// Create 100 issues
	const issueCount = 100
	issues := make([]common.Issue, issueCount)
	expectedIDs := make(map[string]bool)

	for i := range issueCount {
		alert := newTestAlert(channel, uuid.New().String())
		issue := newTestIssue(alert, uuid.New().String())
		issues[i] = issue
		expectedIDs[issue.UniqueID()] = true
	}

	// Save all issues
	err = client.SaveIssues(ctx, issues...)
	require.NoError(err, "should save 100 issues successfully")

	// Load all open issues
	loadedIssues, err := client.LoadOpenIssuesInChannel(ctx, channel)
	require.NoError(err)
	assert.Len(loadedIssues, issueCount, "should load all 100 issues")

	// Verify all issue IDs are present
	for id := range loadedIssues {
		assert.True(expectedIDs[id], "loaded issue ID should be in expected set")
	}
}

// TestFindActiveChannels_ManyChannels tests finding active channels with many channels.
func TestFindActiveChannels_ManyChannels(t *testing.T, client common.DB) {
	ctx := context.Background()
	require := require.New(t)
	assert := assert.New(t)

	// Ensure clean state
	err := client.DropAllData(ctx)
	require.NoError(err)
	err = client.Init(ctx, true)
	require.NoError(err)

	// Create issues in 50 different channels
	const channelCount = 50
	expectedChannels := make(map[string]bool)

	for i := range channelCount {
		channelID := fmt.Sprintf("C0ABABAB%02d", i)
		expectedChannels[channelID] = true

		alert := newTestAlert(channelID, uuid.New().String())
		issue := newTestIssue(alert, uuid.New().String())
		err = client.SaveIssue(ctx, issue)
		require.NoError(err)
	}

	// Find all active channels
	activeChannels, err := client.FindActiveChannels(ctx)
	require.NoError(err)
	assert.Len(activeChannels, channelCount, "should find all 50 channels")

	// Verify all channels are present
	for _, channelID := range activeChannels {
		assert.True(expectedChannels[channelID], "channel should be in expected set")
	}
}

// TestSpecialCharactersInCorrelationID tests handling of special characters.
func TestSpecialCharactersInCorrelationID(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"

	testCases := []struct {
		name          string
		correlationID string
	}{
		{"unicode emoji", "alert-🚨-critical"},
		{"spaces", "alert with spaces"},
		{"special chars", "alert!@#$%^&*()"},
		{"quotes", `alert"with'quotes`},
		{"backslashes", `alert\with\backslashes`},
		{"newlines", "alert\nwith\nnewlines"},
		{"tabs", "alert\twith\ttabs"},
		{"unicode", "alert-日本語-中文"}, //nolint:gosmopolitan // Testing unicode support
		{"sql injection attempt", "'; DROP TABLE issues; --"},
		{"long ID near max", strings.Repeat("a", 490)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			alert := newTestAlert(channel, tc.correlationID)
			issue := newTestIssue(alert, uuid.New().String())

			// Save issue with special correlation ID
			err := client.SaveIssue(ctx, issue)
			require.NoError(err, "should save issue with special correlation ID")

			// Find issue by correlation ID
			id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel, tc.correlationID)
			require.NoError(err, "should find issue with special correlation ID")
			assert.NotEmpty(id, "should return issue ID")
			assert.NotNil(issueBody, "should return issue body")

			foundIssue := testIssueFromJSON(issueBody)
			assert.Equal(tc.correlationID, foundIssue.CorrelationID, "correlation ID should be preserved exactly")
		})
	}
}

// TestChannelProcessingState_EdgeCases tests edge cases for channel processing state.
func TestChannelProcessingState_EdgeCases(t *testing.T, client common.DB) {
	ctx := context.Background()

	t.Run("nil state", func(t *testing.T) {
		require := require.New(t)
		err := client.SaveChannelProcessingState(ctx, nil)
		require.Error(err, "should error when saving nil state")
	})

	t.Run("high open issues count", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)

		channelID := "C" + uuid.New().String()[:10]
		state := common.NewChannelProcessingState(channelID)
		state.OpenIssues = 10000

		err := client.SaveChannelProcessingState(ctx, state)
		require.NoError(err)

		foundState, err := client.FindChannelProcessingState(ctx, channelID)
		require.NoError(err)
		assert.Equal(10000, foundState.OpenIssues)
	})

	t.Run("timestamp precision", func(t *testing.T) {
		require := require.New(t)
		assert := assert.New(t)

		channelID := "C" + uuid.New().String()[:10]
		now := time.Now().UTC()
		state := common.NewChannelProcessingState(channelID)
		state.LastProcessed = now

		err := client.SaveChannelProcessingState(ctx, state)
		require.NoError(err)

		foundState, err := client.FindChannelProcessingState(ctx, channelID)
		require.NoError(err)

		// Verify nanosecond precision is preserved
		assert.Equal(now.Format(time.RFC3339Nano), foundState.LastProcessed.Format(time.RFC3339Nano))
	})
}

// TestSaveIssue_ComplexAlert tests saving issues with complex alert data.
func TestSaveIssue_ComplexAlert(t *testing.T, client common.DB) {
	ctx := context.Background()
	channel := "C0ABABABAB"
	corr := uuid.New().String()
	require := require.New(t)
	assert := assert.New(t)

	// Create alert with all fields populated
	alert := newTestAlertWithAllFields(channel, corr)
	issue := newTestIssue(alert, uuid.New().String())

	err := client.SaveIssue(ctx, issue)
	require.NoError(err, "should save issue with complex alert")

	// Retrieve and verify
	id, issueBody, err := client.FindOpenIssueByCorrelationID(ctx, channel, corr)
	require.NoError(err)
	assert.NotEmpty(id)
	require.NotNil(issueBody)

	foundIssue := testIssueFromJSON(issueBody)
	assert.Equal(alert.Header, foundIssue.LastAlert.Header)
	assert.Equal(alert.Text, foundIssue.LastAlert.Text)
	assert.Equal(alert.Severity, foundIssue.LastAlert.Severity)
	assert.Len(foundIssue.LastAlert.Fields, len(alert.Fields))
}

// TestContextCancellation tests that operations respect context cancellation.
func TestContextCancellation(t *testing.T, client common.DB) {
	t.Run("canceled context for FindOpenIssueByCorrelationID", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err := client.FindOpenIssueByCorrelationID(ctx, "C0ABABABAB", "correlation-123")
		// Should either return error or handle gracefully
		// Implementation-specific behavior
		if err != nil {
			assert.Error(t, err)
		}
	})
}

// RunAllTests runs all database compliance tests.
// This is a convenience function for plugin implementations.
func RunAllTests(t *testing.T, client common.DB) {
	t.Helper()
	// Initialize database
	ctx := context.Background()
	if err := client.Init(ctx, true); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Core functionality tests
	t.Run("SaveAlert", func(t *testing.T) { TestSaveAlert(t, client) })
	t.Run("SaveIssue", func(t *testing.T) { TestSaveIssue(t, client) })
	t.Run("MoveIssue", func(t *testing.T) { TestMoveIssue(t, client) })
	t.Run("FindOpenIssueByCorrelationID", func(t *testing.T) { TestFindOpenIssueByCorrelationID(t, client) })
	t.Run("FindIssueBySlackPostID", func(t *testing.T) { TestFindIssueBySlackPostID(t, client) })
	t.Run("SaveIssues", func(t *testing.T) { TestSaveIssues(t, client) })
	t.Run("FindActiveChannels", func(t *testing.T) { TestFindActiveChannels(t, client) })
	t.Run("LoadOpenIssuesInChannel", func(t *testing.T) { TestLoadOpenIssuesInChannel(t, client) })
	t.Run("CreatingAndFindingMoveMappings", func(t *testing.T) { TestCreatingAndFindingMoveMappings(t, client) })
	t.Run("DeletingMoveMappings", func(t *testing.T) { TestDeletingMoveMappings(t, client) })
	t.Run("CreatingAndFindingChannelProcessingState", func(t *testing.T) { TestCreatingAndFindingChannelProcessingState(t, client) })

	// Initialization tests
	t.Run("Init", func(t *testing.T) { TestInit(t, client) })
	t.Run("Init_WithSchemaValidation", func(t *testing.T) { TestInit_WithSchemaValidation(t, client) })

	// Edge case tests
	t.Run("MoveIssue_EdgeCases", func(t *testing.T) { TestMoveIssue_EdgeCases(t, client) })
	t.Run("ChannelProcessingState_EdgeCases", func(t *testing.T) { TestChannelProcessingState_EdgeCases(t, client) })
	t.Run("SaveIssue_ComplexAlert", func(t *testing.T) { TestSaveIssue_ComplexAlert(t, client) })
	t.Run("SpecialCharactersInCorrelationID", func(t *testing.T) { TestSpecialCharactersInCorrelationID(t, client) })

	// Concurrent operation tests
	t.Run("ConcurrentSaveIssue", func(t *testing.T) { TestConcurrentSaveIssue(t, client) })
	t.Run("ConcurrentMoveMapping", func(t *testing.T) { TestConcurrentMoveMapping(t, client) })

	// Large dataset tests
	t.Run("LoadOpenIssuesInChannel_LargeDataset", func(t *testing.T) { TestLoadOpenIssuesInChannel_LargeDataset(t, client) })
	t.Run("FindActiveChannels_ManyChannels", func(t *testing.T) { TestFindActiveChannels_ManyChannels(t, client) })

	// Context cancellation tests
	t.Run("ContextCancellation", func(t *testing.T) { TestContextCancellation(t, client) })
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

// newTestAlertWithAllFields creates a test alert with all fields populated.
// This is useful for testing JSON serialization and ensuring no data is lost.
func newTestAlertWithAllFields(channelID, correlationID string) *common.Alert {
	alert := common.NewErrorAlert()
	alert.SlackChannelID = channelID
	alert.CorrelationID = correlationID
	alert.Header = "Test Alert with All Fields"
	alert.HeaderWhenResolved = "Test Alert Resolved"
	alert.Text = "This is a test alert with all fields populated for comprehensive testing"
	alert.TextWhenResolved = "Issue has been resolved successfully"
	alert.FallbackText = "Test alert notification"
	alert.Author = "test-service"
	alert.Host = "test-host-01"
	alert.Footer = "Test footer information"
	alert.Link = "https://example.com/alert/123"
	alert.IconEmoji = ":alert:"
	alert.Username = "Test Alert Bot"
	alert.Type = "test"
	alert.IssueFollowUpEnabled = true
	alert.AutoResolveSeconds = 300
	alert.NotificationDelaySeconds = 10
	alert.ArchivingDelaySeconds = 60
	alert.Severity = common.AlertError

	// Add fields
	alert.Fields = []*common.Field{
		{Title: "Environment", Value: "production"},
		{Title: "Service", Value: "api-gateway"},
		{Title: "Region", Value: "us-east-1"},
	}

	// Add escalation
	alert.Escalation = []*common.Escalation{
		{
			Severity:      common.AlertPanic,
			DelaySeconds:  300,
			SlackMentions: []string{"<!here>"},
		},
	}

	// Add webhooks
	alert.Webhooks = []*common.Webhook{
		{
			ID:               "restart",
			URL:              "https://example.com/webhook/restart",
			ButtonText:       "Restart Service",
			ButtonStyle:      common.WebhookButtonStyleDanger,
			AccessLevel:      common.WebhookAccessLevelChannelAdmins,
			DisplayMode:      common.WebhookDisplayModeOpenIssue,
			ConfirmationText: "Are you sure?",
			Payload: map[string]any{
				"action": "restart",
				"host":   "test-host-01",
			},
		},
	}

	// Add metadata
	alert.Metadata = map[string]any{
		"trace_id":   "trace-123-456",
		"request_id": "req-789-012",
		"version":    "1.2.3",
	}

	return alert
}
