// Package handler holds the Echo HTTP handlers for the flag admin API.
package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mralaminahamed/flagcast/packages/shared/bus"
	"github.com/mralaminahamed/flagcast/packages/shared/logger"
	"github.com/mralaminahamed/flagcast/packages/shared/models"
	"github.com/mralaminahamed/flagcast/packages/shared/store"
	"github.com/mralaminahamed/flagcast/packages/shared/validation"
)

// Publisher emits flag-change events so evaluators can refresh their cache.
type Publisher interface {
	PublishJSON(subject string, v any) error
}

type Handler struct {
	store *store.FlagStore
	bus   Publisher // may be nil
	aiURL string    // "" => analysis unavailable
}

func New(s *store.FlagStore, b Publisher, aiURL string) *Handler {
	return &Handler{store: s, bus: b, aiURL: aiURL}
}

type errResponse struct {
	Error string `json:"error"`
}

// flagInput is the create/update request body.
type flagInput struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	Rollout     *int     `json:"rollout"`
	Tags        []string `json:"tags"`
}

// List returns all flags.
func (h *Handler) List(c echo.Context) error {
	flags, err := h.store.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"flags": flags})
}

// Get returns one flag.
func (h *Handler) Get(c echo.Context) error {
	f, err := h.store.Get(c.Request().Context(), c.Param("key"))
	if err != nil {
		return h.storeErr(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// Create adds a flag.
func (h *Handler) Create(c echo.Context) error {
	var in flagInput
	if err := c.Bind(&in); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	in.Key = strings.TrimSpace(in.Key)
	if err := validation.FlagKey(in.Key); err != nil {
		return badRequest(c, err.Error())
	}
	if strings.TrimSpace(in.Name) == "" {
		return badRequest(c, "name is required")
	}
	rollout := 100
	if in.Rollout != nil {
		rollout = *in.Rollout
	}
	if err := validation.Rollout(rollout); err != nil {
		return badRequest(c, err.Error())
	}

	f, err := h.store.Create(c.Request().Context(), models.Flag{
		Key: in.Key, Name: in.Name, Description: in.Description,
		Enabled: in.Enabled, Rollout: rollout, Tags: in.Tags,
	})
	if err != nil {
		return h.storeErr(c, err)
	}
	h.audit(c, f.Key, "created")
	h.notify(f.Key, "created")
	return c.JSON(http.StatusCreated, f)
}

// Update replaces a flag's mutable fields.
func (h *Handler) Update(c echo.Context) error {
	key := c.Param("key")
	var in flagInput
	if err := c.Bind(&in); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	if strings.TrimSpace(in.Name) == "" {
		return badRequest(c, "name is required")
	}
	rollout := 100
	if in.Rollout != nil {
		rollout = *in.Rollout
	}
	if err := validation.Rollout(rollout); err != nil {
		return badRequest(c, err.Error())
	}

	f, err := h.store.Update(c.Request().Context(), models.Flag{
		Key: key, Name: in.Name, Description: in.Description,
		Enabled: in.Enabled, Rollout: rollout, Tags: in.Tags,
	})
	if err != nil {
		return h.storeErr(c, err)
	}
	h.audit(c, key, "updated")
	h.notify(key, "updated")
	return c.JSON(http.StatusOK, f)
}

// Delete removes a flag.
func (h *Handler) Delete(c echo.Context) error {
	key := c.Param("key")
	if err := h.store.Delete(c.Request().Context(), key); err != nil {
		return h.storeErr(c, err)
	}
	h.audit(c, key, "deleted")
	h.notify(key, "deleted")
	return c.NoContent(http.StatusNoContent)
}

// Audit returns the change history, optionally filtered by ?flag=.
func (h *Handler) Audit(c echo.Context) error {
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
	entries, err := h.store.ListAudit(c.Request().Context(), c.QueryParam("flag"), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"audit": entries})
}

// audit records a mutation best-effort; a failure is logged, not surfaced.
func (h *Handler) audit(c echo.Context, key, action string) {
	actor := c.Request().Header.Get("X-Actor")
	if actor == "" {
		actor = "anonymous"
	}
	if err := h.store.AppendAudit(c.Request().Context(), models.AuditEntry{
		FlagKey: key, Action: action, Actor: actor,
	}); err != nil {
		logger.Log.Error().Err(err).Str("flag", key).Msg("append audit")
	}
}

// Analyze forwards an analysis request to the ai service.
func (h *Handler) Analyze(c echo.Context) error {
	if h.aiURL == "" {
		return c.JSON(http.StatusServiceUnavailable, errResponse{"analysis unavailable (AI_URL unset)"})
	}
	resp, err := http.Post(h.aiURL+"/analyze", "application/json", c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadGateway, errResponse{err.Error()})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return c.Blob(resp.StatusCode, "application/json", body)
}

// notify publishes a flag-change event best-effort so evaluators refresh.
func (h *Handler) notify(key, action string) {
	if h.bus == nil {
		return
	}
	if err := h.bus.PublishJSON(bus.SubjectFlagChanged, bus.FlagChanged{Key: key, Action: action}); err != nil {
		logger.Log.Error().Err(err).Str("flag", key).Msg("publish flag.changed")
	}
}

func (h *Handler) storeErr(c echo.Context, err error) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return c.JSON(http.StatusNotFound, errResponse{err.Error()})
	case errors.Is(err, store.ErrConflict):
		return c.JSON(http.StatusConflict, errResponse{err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, errResponse{err.Error()})
	}
}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, errResponse{msg})
}
