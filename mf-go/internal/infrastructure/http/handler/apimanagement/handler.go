package apimanagement

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/apimanagement/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/apimanagement/usecase"
	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/pagination"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
	"github.com/masterfabric-go/masterfabric/internal/shared/validator"
)

// Handler provides API Management HTTP handlers.
type Handler struct {
	defineEndpointUC  *usecase.DefineEndpointUseCase
	updatePolicyUC    *usecase.UpdatePolicyUseCase
	retireEndpointUC  *usecase.RetireEndpointUseCase
	activateEndpointUC *usecase.ActivateEndpointUseCase
	endpointRepo      repository.EndpointRepository
	policyRepo        repository.PolicyRepository
}

// NewHandler creates a new API Management handler.
func NewHandler(
	defineEndpointUC *usecase.DefineEndpointUseCase,
	updatePolicyUC *usecase.UpdatePolicyUseCase,
	retireEndpointUC *usecase.RetireEndpointUseCase,
	activateEndpointUC *usecase.ActivateEndpointUseCase,
	endpointRepo repository.EndpointRepository,
	policyRepo repository.PolicyRepository,
) *Handler {
	return &Handler{
		defineEndpointUC:  defineEndpointUC,
		updatePolicyUC:    updatePolicyUC,
		retireEndpointUC:  retireEndpointUC,
		activateEndpointUC: activateEndpointUC,
		endpointRepo:      endpointRepo,
		policyRepo:        policyRepo,
	}
}

// DefineEndpoint creates a new endpoint.
func (h *Handler) DefineEndpoint(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "appId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid app id"})
		return
	}

	var req dto.DefineEndpointRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ep, err := h.defineEndpointUC.Execute(r.Context(), appID, req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, ep)
}

// GetEndpoint returns an endpoint by ID.
func (h *Handler) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endpoint id"})
		return
	}

	ep, err := h.endpointRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto.EndpointInfo{
		ID:             ep.ID,
		AppID:          ep.AppID,
		Method:         ep.Method,
		Path:           ep.Path,
		Version:        ep.Version,
		BackendService: ep.BackendService,
		BackendAction:  ep.BackendAction,
		Schema:         ep.Schema,
		AuditLevel:     ep.AuditLevel,
		PIIMasking:     ep.PIIMasking,
		EventAfter:     ep.EventAfter,
		Status:         string(ep.Status),
		CreatedAt:      ep.CreatedAt,
	})
}

// ListEndpoints returns endpoints for an app.
func (h *Handler) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "appId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid app id"})
		return
	}

	params := pagination.FromRequest(r)
	endpoints, total, err := h.endpointRepo.ListByApp(r.Context(), appID, params.Offset(), params.Limit())
	if err != nil {
		response.Error(w, err)
		return
	}

	var infos []dto.EndpointInfo
	for _, ep := range endpoints {
		infos = append(infos, dto.EndpointInfo{
			ID:             ep.ID,
			AppID:          ep.AppID,
			Method:         ep.Method,
			Path:           ep.Path,
			Version:        ep.Version,
			BackendService: ep.BackendService,
			BackendAction:  ep.BackendAction,
			AuditLevel:     ep.AuditLevel,
			PIIMasking:     ep.PIIMasking,
			Status:         string(ep.Status),
			CreatedAt:      ep.CreatedAt,
		})
	}

	response.JSON(w, http.StatusOK, pagination.NewResult(infos, params, total))
}

// RetireEndpoint retires an endpoint.
func (h *Handler) RetireEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endpoint id"})
		return
	}

	if err := h.retireEndpointUC.Execute(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

// ActivateEndpoint activates a retired or inactive endpoint.
func (h *Handler) ActivateEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endpoint id"})
		return
	}

	if err := h.activateEndpointUC.Execute(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

// UpdatePolicy updates an endpoint's policy.
func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	endpointID, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endpoint id"})
		return
	}

	var req dto.UpdatePolicyRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	policy, err := h.updatePolicyUC.Execute(r.Context(), endpointID, req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, policy)
}

// GetPolicy returns the policy for an endpoint.
func (h *Handler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	endpointID, err := uuid.Parse(chi.URLParam(r, "endpointId"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endpoint id"})
		return
	}

	policy, err := h.policyRepo.GetByEndpointID(r.Context(), endpointID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto.PolicyInfo{
		ID:                 policy.ID,
		EndpointID:         policy.EndpointID,
		RequiredPermission: policy.RequiredPermission,
		RateLimit:          policy.RateLimit,
		AuthPolicy:         policy.AuthPolicy,
		ValidationPolicy:   policy.ValidationPolicy,
		ExtraPolicies:      policy.ExtraPolicies,
	})
}
