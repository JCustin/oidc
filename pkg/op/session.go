package op

import (
	"context"
	"net/http"

	"github.com/caos/oidc/pkg/oidc"
	"github.com/caos/oidc/pkg/rp"
	"github.com/gorilla/schema"
)

type SessionEnder interface {
	Decoder() *schema.Decoder
	Storage() Storage
	IDTokenVerifier() rp.Verifier
	DefaultLogoutRedirectURI() string
}

func EndSession(w http.ResponseWriter, r *http.Request, ender SessionEnder) {
	req, err := ParseEndSessionRequest(r, ender.Decoder())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, err := ValidateEndSessionRequest(r.Context(), req, ender)
	if err != nil {
		RequestError(w, r, err)
		return
	}
	var clientID string
	if session.Client != nil {
		clientID = session.Client.GetID()
	}
	err = ender.Storage().TerminateSession(r.Context(), session.UserID, clientID)
	if err != nil {
		RequestError(w, r, ErrServerError("error terminating session"))
		return
	}
	http.Redirect(w, r, session.RedirectURI, http.StatusFound)
}

func ParseEndSessionRequest(r *http.Request, decoder *schema.Decoder) (*oidc.EndSessionRequest, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, ErrInvalidRequest("error parsing form")
	}
	req := new(oidc.EndSessionRequest)
	err = decoder.Decode(req, r.Form)
	if err != nil {
		return nil, ErrInvalidRequest("error decoding form")
	}
	return req, nil
}

func ValidateEndSessionRequest(ctx context.Context, req *oidc.EndSessionRequest, ender SessionEnder) (*EndSessionRequest, error) {
	session := new(EndSessionRequest)
	if req.IdTokenHint == "" {
		return session, nil
	}
	claims, err := ender.IDTokenVerifier().Verify(ctx, "", req.IdTokenHint)
	if err != nil {
		return nil, ErrInvalidRequest("id_token_hint invalid")
	}
	session.UserID = claims.Subject
	session.Client, err = ender.Storage().GetClientByClientID(ctx, claims.AuthorizedParty)
	if err != nil {
		return nil, ErrServerError("")
	}
	if req.PostLogoutRedirectURI == "" {
		session.RedirectURI = ender.DefaultLogoutRedirectURI()
		return session, nil
	}
	for _, uri := range session.Client.PostLogoutRedirectURIs() {
		if uri == req.PostLogoutRedirectURI {
			session.RedirectURI = uri + "?state=" + req.State
			return session, nil
		}
	}
	return nil, ErrInvalidRequest("post_logout_redirect_uri invalid")
}

func NeedsExistingSession(authRequest *oidc.AuthRequest) bool {
	if authRequest == nil {
		return true
	}
	if authRequest.Prompt == oidc.PromptNone {
		return true
	}
	return false
}
