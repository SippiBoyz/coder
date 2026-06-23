package aibridgedserver_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cdr.dev/slog/v3/sloggers/slogtest"
	"github.com/coder/coder/v2/coderd/aibridged/proto"
	"github.com/coder/coder/v2/coderd/aibridgedserver"
	agplaiseats "github.com/coder/coder/v2/coderd/aiseats"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbgen"
	"github.com/coder/coder/v2/coderd/database/dbtestutil"
	"github.com/coder/coder/v2/coderd/util/ptr"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/testutil"
)

// TestGetAIProviders exercises the row-to-proto mapping over a real database:
// enabled providers carry their keys (and typed Bedrock settings), disabled
// providers are included but withhold keys and settings, and Copilot (a
// keyless BYOK provider) round-trips with no keys.
func TestGetAIProviders(t *testing.T) {
	t.Parallel()

	db, _ := dbtestutil.NewDB(t)
	ctx := testutil.Context(t, testutil.WaitLong)
	logger := slogtest.Make(t, nil)

	// Enabled OpenAI with two keys.
	openai := dbgen.AIProvider(t, db, database.AIProvider{
		Type:    database.AIProviderTypeOpenai,
		Name:    "openai",
		Enabled: true,
		BaseUrl: "https://api.openai.com/",
	})
	dbgen.AIProviderKey(t, db, database.AIProviderKey{ProviderID: openai.ID, APIKey: "sk-openai-1"})
	dbgen.AIProviderKey(t, db, database.AIProviderKey{ProviderID: openai.ID, APIKey: "sk-openai-2"})

	// Enabled Bedrock with typed settings.
	bedrockSettings, err := json.Marshal(codersdk.AIProviderSettings{
		Bedrock: &codersdk.AIProviderBedrockSettings{
			Region:          "us-east-1",
			Model:           "anthropic.claude-3",
			SmallFastModel:  "anthropic.claude-haiku",
			AccessKey:       ptr.Ref("AKID"),
			AccessKeySecret: ptr.Ref("secret"),
		},
	})
	require.NoError(t, err)
	dbgen.AIProvider(t, db, database.AIProvider{
		Type:     database.AIProviderTypeBedrock,
		Name:     "bedrock",
		Enabled:  true,
		BaseUrl:  "https://bedrock-runtime.us-east-1.amazonaws.com/",
		Settings: sql.NullString{String: string(bedrockSettings), Valid: true},
	})

	// Enabled Copilot, which is keyless (BYOK per request).
	dbgen.AIProvider(t, db, database.AIProvider{
		Type:    database.AIProviderTypeCopilot,
		Name:    "copilot",
		Enabled: true,
		BaseUrl: "https://api.githubcopilot.com/",
	})

	// Disabled Anthropic with a key; the key must be withheld.
	disabled := dbgen.AIProvider(t, db, database.AIProvider{
		Type:    database.AIProviderTypeAnthropic,
		Name:    "anthropic-off",
		BaseUrl: "https://api.anthropic.com/",
	}, func(p *database.InsertAIProviderParams) {
		p.Enabled = false
	})
	dbgen.AIProviderKey(t, db, database.AIProviderKey{ProviderID: disabled.ID, APIKey: "sk-secret"})

	srv, err := aibridgedserver.NewServer(ctx, db, logger, "/", codersdk.AIBridgeConfig{}, nil, nil, agplaiseats.Noop{})
	require.NoError(t, err)

	resp, err := srv.GetAIProviders(ctx, &proto.GetAIProvidersRequest{})
	require.NoError(t, err)

	byName := make(map[string]*proto.AIProvider, len(resp.GetProviders()))
	for _, p := range resp.GetProviders() {
		byName[p.GetName()] = p
	}
	require.Len(t, byName, 4)

	gotOpenAI := byName["openai"]
	require.NotNil(t, gotOpenAI)
	assert.True(t, gotOpenAI.GetEnabled())
	assert.Equal(t, string(database.AIProviderTypeOpenai), gotOpenAI.GetType())
	assert.Equal(t, "https://api.openai.com/", gotOpenAI.GetBaseUrl())
	assert.ElementsMatch(t, []string{"sk-openai-1", "sk-openai-2"}, gotOpenAI.GetKeys())
	assert.Nil(t, gotOpenAI.GetBedrock())

	gotBedrock := byName["bedrock"]
	require.NotNil(t, gotBedrock)
	assert.True(t, gotBedrock.GetEnabled())
	require.NotNil(t, gotBedrock.GetBedrock())
	assert.Equal(t, "us-east-1", gotBedrock.GetBedrock().GetRegion())
	assert.Equal(t, "anthropic.claude-3", gotBedrock.GetBedrock().GetModel())
	assert.Equal(t, "anthropic.claude-haiku", gotBedrock.GetBedrock().GetSmallFastModel())
	assert.Equal(t, "AKID", gotBedrock.GetBedrock().GetAccessKey())
	assert.Equal(t, "secret", gotBedrock.GetBedrock().GetAccessKeySecret())

	gotCopilot := byName["copilot"]
	require.NotNil(t, gotCopilot)
	assert.True(t, gotCopilot.GetEnabled())
	assert.Empty(t, gotCopilot.GetKeys())

	gotDisabled := byName["anthropic-off"]
	require.NotNil(t, gotDisabled)
	assert.False(t, gotDisabled.GetEnabled())
	assert.Empty(t, gotDisabled.GetKeys(), "keys must be withheld for disabled providers")
	assert.Nil(t, gotDisabled.GetBedrock())
}

// TestGetAIProvidersBlocksOnSeedLock asserts that GetAIProviders serializes on
// LockIDAIProvidersEnvSeed: while an in-flight seed transaction holds the lock,
// the fetch blocks, and once the seed commits the fetch returns the seeded
// set. Postgres advisory locks are required, so this cannot run against the
// mock store.
func TestGetAIProvidersBlocksOnSeedLock(t *testing.T) {
	t.Parallel()

	db, _ := dbtestutil.NewDB(t)
	ctx := testutil.Context(t, testutil.WaitLong)
	logger := slogtest.Make(t, nil)

	dbgen.AIProviderWithOptionalKey(t, db, database.AIProvider{
		Type:    database.AIProviderTypeOpenai,
		Name:    "openai",
		Enabled: true,
		BaseUrl: "https://api.openai.com/",
	}, "sk-openai")

	srv, err := aibridgedserver.NewServer(ctx, db, logger, "/", codersdk.AIBridgeConfig{}, nil, nil, agplaiseats.Noop{})
	require.NoError(t, err)

	// Simulate an in-flight env seed holding the advisory lock until released.
	holderReady := make(chan struct{})
	releaseHolder := make(chan struct{})
	holderDone := make(chan struct{})
	go func() {
		defer close(holderDone)
		txErr := db.InTx(func(tx database.Store) error {
			if err := tx.AcquireLock(ctx, database.LockIDAIProvidersEnvSeed); err != nil {
				return err
			}
			close(holderReady)
			<-releaseHolder
			return nil
		}, nil)
		assert.NoError(t, txErr)
	}()

	testutil.TryReceive(ctx, t, holderReady)

	fetchDone := make(chan *proto.GetAIProvidersResponse, 1)
	fetchErr := make(chan error, 1)
	go func() {
		resp, err := srv.GetAIProviders(ctx, &proto.GetAIProvidersRequest{})
		fetchErr <- err
		fetchDone <- resp
	}()

	// While the seed lock is held, the fetch must not complete.
	select {
	case <-fetchDone:
		t.Fatal("GetAIProviders returned before the seed lock was released")
	case <-time.After(testutil.IntervalMedium):
	}

	// Release the lock; the fetch should now complete and return the seeded set.
	close(releaseHolder)
	testutil.TryReceive(ctx, t, holderDone)

	require.NoError(t, testutil.TryReceive(ctx, t, fetchErr))
	resp := testutil.TryReceive(ctx, t, fetchDone)
	require.Len(t, resp.GetProviders(), 1)
	assert.Equal(t, "openai", resp.GetProviders()[0].GetName())
	assert.Equal(t, []string{"sk-openai"}, resp.GetProviders()[0].GetKeys())
}
