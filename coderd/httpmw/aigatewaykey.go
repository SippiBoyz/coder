package httpmw

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/coder/coder/v2/coderd/apikey"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/dbauthz"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/codersdk"
)

type aiGatewayKeyContextKey struct{}

// AIGatewayKeyAuthOptional returns the ID of the AI Gateway key that
// authenticated the request, if any. The /serve handler uses it to record
// liveness against the authenticating key.
func AIGatewayKeyAuthOptional(r *http.Request) (uuid.UUID, bool) {
	id, ok := r.Context().Value(aiGatewayKeyContextKey{}).(uuid.UUID)
	return id, ok
}

// ExtractAIGatewayKeyConfig configures ExtractAIGatewayKeyAuthenticated.
type ExtractAIGatewayKeyConfig struct {
	DB database.Store
	// Optional, when true, allows the request to proceed unauthenticated. The
	// next handler can detect authentication via AIGatewayKeyAuthOptional.
	Optional bool
}

// ExtractAIGatewayKeyAuthenticated authenticates a request as a standalone AI
// Gateway replica using the X-AI-Governance-Gateway-Key header. The header
// value is hashed and looked up in the ai_gateway_keys table, mirroring the
// external provisioner daemon key flow in ExtractProvisionerDaemonAuthenticated.
//
// One key may authenticate many replicas at once; keys are not unique per
// connection. A deleted key fails the next reconnect because the row no longer
// exists, but does not disconnect sessions already established.
func ExtractAIGatewayKeyAuthenticated(opts ExtractAIGatewayKeyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			handleOptional := func(code int, response codersdk.Response) {
				if opts.Optional {
					next.ServeHTTP(w, r)
					return
				}
				httpapi.Write(ctx, w, code, response)
			}

			key := r.Header.Get(codersdk.AIGatewayKeyHeader)
			if key == "" {
				handleOptional(http.StatusUnauthorized, codersdk.Response{
					Message: "AI Gateway key required.",
				})
				return
			}

			hashedKey := apikey.HashSecret(key)
			// nolint:gocritic // System must look up the AI Gateway key to authenticate the request.
			keyID, err := opts.DB.GetAIGatewayKeyIDByHashedSecret(dbauthz.AsSystemRestricted(ctx), hashedKey)
			if err != nil {
				// The lookup is an exact match on a unique index, so a missing
				// row means the key is invalid or revoked.
				if httpapi.Is404Error(err) {
					handleOptional(http.StatusUnauthorized, codersdk.Response{
						Message: "AI Gateway key invalid.",
					})
					return
				}
				handleOptional(http.StatusInternalServerError, codersdk.Response{
					Message: "Failed to look up AI Gateway key.",
					Detail:  err.Error(),
				})
				return
			}

			ctx = context.WithValue(ctx, aiGatewayKeyContextKey{}, keyID)
			// nolint:gocritic // Authenticating as an AI Gateway replica, which
			// acts as the AI Bridge daemon.
			ctx = dbauthz.AsAIBridged(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
