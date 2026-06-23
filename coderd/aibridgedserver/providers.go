package aibridgedserver

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"cdr.dev/slog/v3"
	"github.com/coder/coder/v2/coderd/aibridged/proto"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/db2sdk"
	"github.com/coder/coder/v2/coderd/database/dbauthz"
	"github.com/coder/coder/v2/coderd/util/ptr"
)

// GetAIProviders returns the full AI provider set (enabled and disabled) from
// the database, which is the single source of truth seeded from coderd's
// environment. Embedded and standalone AI Gateway daemons call this over DRPC
// to build their provider pool instead of reading the database directly.
//
// The handler reads under a read-only transaction that first acquires
// LockIDAIProvidersEnvSeed, so it blocks until any in-flight env seed commits
// or rolls back. This guarantees the response is never a partial, mid-seed
// snapshot.
//
// Keys are populated only for enabled providers; disabled providers never call
// upstream, so their secrets are withheld.
//
// SECURITY: the response carries plaintext API keys and Bedrock credentials.
// Do not log the response struct.
func (s *Server) GetAIProviders(ctx context.Context, _ *proto.GetAIProvidersRequest) (*proto.GetAIProvidersResponse, error) {
	//nolint:gocritic // AIBridged has a minimal permission set for this purpose, matching the embedded DB-read path.
	ctx = dbauthz.AsAIBridged(ctx)

	var (
		rows           []database.AIProvider
		keysByProvider map[uuid.UUID][]database.AIProviderKey
	)
	// Wrap both reads in a read-only transaction so the provider list and the
	// key list are consistent with each other, and so the seed lock is held
	// for the duration of the reads.
	err := s.store.InTx(func(tx database.Store) error {
		// Block on any in-flight seed transaction holding the advisory lock so
		// the response reflects a fully-seeded snapshot.
		if err := tx.AcquireLock(ctx, database.LockIDAIProvidersEnvSeed); err != nil {
			return xerrors.Errorf("acquire ai providers env seed lock: %w", err)
		}

		var err error
		rows, err = tx.GetAIProviders(ctx, database.GetAIProvidersParams{IncludeDisabled: true})
		if err != nil {
			return xerrors.Errorf("get ai providers: %w", err)
		}

		// Load keys only for enabled providers to avoid materializing secrets
		// for disabled rows.
		ids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			if !row.Enabled {
				continue
			}
			ids = append(ids, row.ID)
		}
		keysByProvider = make(map[uuid.UUID][]database.AIProviderKey, len(ids))
		if len(ids) == 0 {
			return nil
		}
		keyRows, err := tx.GetAIProviderKeysByProviderIDs(ctx, ids)
		if err != nil {
			return xerrors.Errorf("get ai provider keys: %w", err)
		}
		for _, k := range keyRows {
			keysByProvider[k.ProviderID] = append(keysByProvider[k.ProviderID], k)
		}
		return nil
	}, &database.TxOptions{ReadOnly: true, TxIdentifier: "get_ai_providers"})
	if err != nil {
		return nil, err
	}

	providers := make([]*proto.AIProvider, 0, len(rows))
	for _, row := range rows {
		p, err := aiProviderToProto(row, keysByProvider[row.ID])
		if err != nil {
			// Skip the offending row rather than failing the whole fetch:
			// one row with a corrupt settings blob must not break provider
			// configuration for every gateway, which would otherwise loop
			// forever on the empty pool.
			s.logger.Error(ctx, "skipping ai provider that failed to map",
				slog.F("provider_id", row.ID),
				slog.F("provider_name", row.Name),
				slog.F("provider_type", string(row.Type)),
				slog.Error(err),
			)
			continue
		}
		providers = append(providers, p)
	}

	return &proto.GetAIProvidersResponse{Providers: providers}, nil
}

// aiProviderToProto maps a single ai_providers row (and its keys, for enabled
// providers) to the proto representation served to AI Gateway daemons. Keys and
// Bedrock settings are only attached for enabled providers; disabled providers
// never call upstream so their secrets are withheld.
func aiProviderToProto(row database.AIProvider, keys []database.AIProviderKey) (*proto.AIProvider, error) {
	p := &proto.AIProvider{
		Name:    row.Name,
		Type:    string(row.Type),
		Enabled: row.Enabled,
		BaseUrl: row.BaseUrl,
	}
	if !row.Enabled {
		return p, nil
	}

	p.Keys = make([]string, 0, len(keys))
	for _, k := range keys {
		p.Keys = append(p.Keys, k.APIKey)
	}

	settings, err := db2sdk.AIProviderSettings(row.Settings)
	if err != nil {
		return nil, xerrors.Errorf("decode settings: %w", err)
	}
	if settings.Bedrock != nil {
		p.Bedrock = &proto.AIProviderBedrock{
			Region:          settings.Bedrock.Region,
			AccessKey:       ptr.NilToEmpty(settings.Bedrock.AccessKey),
			AccessKeySecret: ptr.NilToEmpty(settings.Bedrock.AccessKeySecret),
			Model:           settings.Bedrock.Model,
			SmallFastModel:  settings.Bedrock.SmallFastModel,
		}
	}

	return p, nil
}
