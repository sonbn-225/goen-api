package v1
 
import (
	"encoding/json"
	"net/http"
	"strconv"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)
 
type AccountHandler struct {
	svc interfaces.AccountService
}
 
func NewAccountHandler(svc interfaces.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}
 
func (h *AccountHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
 
		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Route("/{accountId}", func(r chi.Router) {
				r.Get("/", h.Get)
				r.Patch("/", h.Patch)
				r.Delete("/", h.Delete)
				r.Get("/shares", h.ListShares)
				r.Get("/audit-events", h.ListAuditEvents)
				r.Put("/shares", h.UpsertShare)
				r.Delete("/shares/{userId}", h.RevokeShare)
			})
		})
	})
}
 
// List godoc
// @Summary List Accounts
// @Description Retrieve a list of accounts for the current user
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	// 1. Trích xuất ID của người dùng đang đăng nhập từ Context (được gài vào bởi AuthMiddleware)
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Gọi Service để lấy danh sách tất cả các tài khoản thuộc sở hữu (hoặc được share) của người dùng
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	// 3. Trả về kết quả JSON với HTTP status 200 (OK)
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// Create godoc
// @Summary Create Account
// @Description Create a new banking or manual account for the user
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAccountRequest true "Account Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Xác thực người dùng đang request
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Decode body của request (JSON) vào DTO CreateAccountRequest
	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}
 
	// 3. Gọi Service để thực hiện logic tạo tài khoản (kiểm tra type, khởi tạo số dư, lưu DB)
	account, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	// 4. Trả về thông tin tài khoản vừa tạo với status 201 (Created)
	response.WriteSuccess(w, http.StatusCreated, account)
}
 
// Get godoc
// @Summary Get Account
// @Description Retrieve specific account properties by ID
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [get]
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	// 1. Xác thực User
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Lấy AccountID từ đường dẫn URL (ex: /accounts/12345) và parse sang UUID
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	// 3. Query Service lấy chi tiết tài khoản (Service sẽ kiểm tra cả quyền truy cập - ownership/share)
	acc, err := h.svc.Get(r.Context(), userID, accountID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	if acc == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "account not found", nil)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, acc)
}
 
// Patch godoc
// @Summary Update Account
// @Description Partially update account information
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body dto.PatchAccountRequest true "Account Patch Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AccountResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [patch]
func (h *AccountHandler) Patch(w http.ResponseWriter, r *http.Request) {
	// 1. Xác định User đang request
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Tìm ID của tài khoản cần update
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	// 3. Bind JSON payload vào request dto. Những field nil sẽ bị bỏ qua khi update (Patch semantics)
	var req dto.PatchAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}
 
	// 4. Áp dụng thay đổi và ghi nhận lịch sử thay đổi (Audit event) bên trong Service
	acc, err := h.svc.Patch(r.Context(), userID, accountID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	if acc == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "account not found", nil)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, acc)
}
 
// Delete godoc
// @Summary Delete Account
// @Description Delete an account (and its associated dependencies according to business rules)
// @Tags Accounts
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. Xác minh định danh user
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Parse UUID tài khoản cần xóa
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	// 3. Gọi hàm xóa của Service (sẽ là Soft Delete hoặc chặn lại nếu tài khoản đang có số dư/giao dịch)
	if err := h.svc.Delete(r.Context(), userID, accountID); err != nil {
		response.HandleError(w, err)
		return
	}
 
	// 4. Trả về mã 204 tượng trưng cho trạng thái Delete thành công, không có Body.
	w.WriteHeader(http.StatusNoContent)
}
 
// ListShares godoc
// @Summary List Account Shares
// @Description List active sharing links provided to other users for this account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountShareResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares [get]
func (h *AccountHandler) ListShares(w http.ResponseWriter, r *http.Request) {
	// 1. Phân quyền API
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	// 2. Lấy Account ID từ URL Parameter
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	// 3. Danh sách những ai đang được chia sẻ quyền truy cập tài khoản này (chỉ chủ sở hữu mới xem được)
	items, err := h.svc.ListShares(r.Context(), userID, accountID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// ListAuditEvents godoc
// @Summary List Account Audit Events
// @Description List recent audit events for an account visible to the current user
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param limit query int false "Max number of events" default(50)
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.AccountAuditEventResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/audit-events [get]
func (h *AccountHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "limit must be an integer", nil)
			return
		}
		limit = parsedLimit
	}
 
	items, err := h.svc.ListAuditEvents(r.Context(), userID, accountID, limit)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// UpsertShare godoc
// @Summary Upsert Account Share
// @Description Add or update view/admin permissions for another user on an account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body dto.UpsertShareRequest true "Upsert Share Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AccountShareResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares [put]
func (h *AccountHandler) UpsertShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	var req dto.UpsertShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}
 
	item, err := h.svc.UpsertShare(r.Context(), userID, accountID, req.Login, req.Permission)
	if err != nil {
		response.HandleError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, item)
}
 
// RevokeShare godoc
// @Summary Revoke Account Share
// @Description Remove access permissions given to another user for this account
// @Tags Accounts
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param userId path string true "Target User ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Router /accounts/{accountId}/shares/{userId} [delete]
func (h *AccountHandler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
 
	accountID, err := uuid.Parse(chi.URLParam(r, "accountId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
 
	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid target user id format", nil)
		return
	}
 
	if err := h.svc.RevokeShare(r.Context(), userID, accountID, targetUserID); err != nil {
		response.HandleError(w, err)
		return
	}
 
	w.WriteHeader(http.StatusNoContent)
}
