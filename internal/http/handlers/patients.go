package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cprt-lis/internal/domain"

	"github.com/labstack/echo/v4"
)

type PatientHandler struct {
	service PatientService
}

func NewPatientHandler(service PatientService) *PatientHandler {
	return &PatientHandler{service: service}
}

// CreatePatientRequest is the payload for POST /api/patients/register (and /api/patients).
type CreatePatientRequest struct {
	// Title / salutation
	Prefix string `json:"prefix"`
	// Patient first (and last) name — required
	FirstName string `json:"first_name"`
	// M, F or Other
	Gender string `json:"gender"`
	// Non-negative integer
	Age int `json:"age"`
	// Yrs | Mon | Days
	AgeUnit string `json:"age_unit"`
	// Primary phone field
	PhoneNo string `json:"phone_no"`
	// Alias for phone_no — either field is accepted
	Phone string `json:"phone"`
	// OP/IP registration number from the hospital
	OPIPNo string `json:"op_ip_no"`
	// Outpatient | Inpatient
	PatientType string `json:"patient_type"`
	CreatedBy   string `json:"created_by" `
	Status      string `json:"status"     `
}

// PatientRegisterResponse is each element of the array returned after a patient is registered.
type PatientRegisterResponse struct {
	// UUID assigned to the patient
	PatientID        string    `json:"patient_id"        `
	PatientNo        int64     `json:"patient_no"        `
	FullName         string    `json:"full_name"         `
	RegistrationDate string    `json:"registration_date" `
	Status           string    `json:"status"            `
	CreatedAt        time.Time `json:"created_at"`
}

// Register registers a new patient.
//
//	@Summary      Register patient
//	@Description  Creates a new patient record and returns the assigned patient_id (UUID) and patient_no.
//	@Description  age_unit accepts: Yrs, Mon, Days (and common aliases like years, months, days).
//	@Tags         Patients
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreatePatientRequest    true  "Patient details"
//	@Success      201   {array}   PatientRegisterResponse "Single-element array with the created patient record"
//	@Failure      400   {object}  ErrorResponse           "Validation error — e.g. first_name is required, invalid age_unit"
//	@Failure      401   {object}  ErrorResponse           "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse           "Unexpected server error"
//	@Router       /api/patients/register [post]
func (h *PatientHandler) Register(c echo.Context) error {
	var req CreatePatientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if strings.TrimSpace(req.FirstName) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "first_name is required")
	}
	if req.Age < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "age must be greater than or equal to 0")
	}

	ageUnit, err := normalizeAgeUnit(req.AgeUnit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	phoneNo := strings.TrimSpace(req.PhoneNo)
	if phoneNo == "" {
		phoneNo = strings.TrimSpace(req.Phone)
	}

	patient, err := h.service.Register(c.Request().Context(), domain.Patient{
		Prefix:      strings.TrimSpace(req.Prefix),
		FirstName:   strings.TrimSpace(req.FirstName),
		Gender:      strings.TrimSpace(req.Gender),
		Age:         req.Age,
		AgeUnit:     ageUnit,
		Phone:       phoneNo,
		OPIPNo:      strings.TrimSpace(req.OPIPNo),
		PatientType: strings.TrimSpace(req.PatientType),
		CreatedBy:   strings.TrimSpace(req.CreatedBy),
		Status:      strings.TrimSpace(req.Status),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	fullName := strings.TrimSpace(strings.TrimSpace(patient.Prefix) + " " + strings.TrimSpace(patient.FirstName))
	if fullName == "" {
		fullName = strings.TrimSpace(patient.FirstName)
	}
	if patient.Status == "" {
		patient.Status = "Active"
	}

	resp := []PatientRegisterResponse{{
		PatientID:        patient.PatientUUID,
		PatientNo:        patient.PatientNo,
		FullName:         fullName,
		RegistrationDate: patient.CreatedAt.Format("2006-01-02"),
		Status:           patient.Status,
		CreatedAt:        patient.CreatedAt,
	}}

	return c.JSON(http.StatusCreated, resp)
}

// Create is an alias for Register — POST /api/patients.
//
//	@Summary      Create patient (alias)
//	@Description  Identical to POST /api/patients/register. Kept for backward compatibility.
//	@Tags         Patients
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreatePatientRequest    true  "Patient details"
//	@Success      201   {array}   PatientRegisterResponse "Single-element array with the created patient record"
//	@Failure      400   {object}  ErrorResponse           "Validation error"
//	@Failure      401   {object}  ErrorResponse           "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse           "Unexpected server error"
//	@Router       /api/patients [post]
func (h *PatientHandler) Create(c echo.Context) error {
	return h.Register(c)
}

// GetByID returns the visit/test history for a patient.
//
//	@Summary      Get patient history by UUID
//	@Description  Returns an array of PatientHistoryItem records for the given patient UUID.
//	@Tags         Patients
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string                   true  "Patient UUID" example(550e8400-e29b-41d4-a716-446655440000)
//	@Success      200  {array}   domain.PatientHistoryItem "Bill-level history entries"
//	@Failure      400  {object}  ErrorResponse             "patient id is required"
//	@Failure      401  {object}  ErrorResponse             "Missing or invalid JWT"
//	@Failure      500  {object}  ErrorResponse             "Unexpected server error"
//	@Router       /api/patients/{id} [get]
func (h *PatientHandler) GetByID(c echo.Context) error {
	patientUUID := strings.TrimSpace(c.Param("id"))
	if patientUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient id is required")
	}

	history, err := h.service.History(c.Request().Context(), patientUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, history)
}

// Search searches for patients by name, phone, or OP/IP number.
//
//	@Summary      Search patients
//	@Description  Full-text search across patient name, phone, and OP/IP no. Returns matching PatientSearchResult records.
//	@Description  Pass the search term as ?query= or ?search= — both are accepted.
//	@Tags         Patients
//	@Produce      json
//	@Security     BearerAuth
//	@Param        query  query     string                     false  "Free-text search term" example(Rahul)
//	@Param        search query     string                     false  "Alias for query param"  example(Rahul)
//	@Success      200    {array}   domain.PatientSearchResult "Matched patients (empty array if none found)"
//	@Failure      401    {object}  ErrorResponse              "Missing or invalid JWT"
//	@Failure      500    {object}  ErrorResponse              "Unexpected server error"
//	@Router       /api/patients [get]
func (h *PatientHandler) Search(c echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("query"))
	if query == "" {
		query = strings.TrimSpace(c.QueryParam("search"))
	}

	results, err := h.service.Search(c.Request().Context(), query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, results)
}

// UpdatePatientRequest is the payload for PATCH /api/patients/:patientId.
// All fields are optional — only supplied fields are updated.
type UpdatePatientRequest struct {
	Prefix    string `json:"prefix"      `
	FirstName string `json:"first_name"  `
	// M, F or Other
	Gender string `json:"gender"      `
	Age    int    `json:"age"         `
	// Yrs | Mon | Days
	AgeUnit string `json:"age_unit"    `
	PhoneNo string `json:"phone_no"    `
	// Outpatient | Inpatient
	PatientType string `json:"patient_type"`
	// Active | Inactive
	Status    string `json:"status"      `
	UpdatedBy string `json:"updated_by"  `
}

// GetHistory returns the test/billing history for a patient.
//
//	@Summary      Get patient history
//	@Description  Returns an array of PatientHistoryItem records (one per service/test on each bill) for the given patient.
//	@Tags         Patients
//	@Produce      json
//	@Security     BearerAuth
//	@Param        patientId  path      string                   true  "Patient UUID" example(550e8400-e29b-41d4-a716-446655440000)
//	@Success      200        {array}   domain.PatientHistoryItem "History entries — bill date, service, status and rate"
//	@Failure      400        {object}  ErrorResponse             "patient id is required"
//	@Failure      401        {object}  ErrorResponse             "Missing or invalid JWT"
//	@Failure      500        {object}  ErrorResponse             "Unexpected server error"
//	@Router       /api/patients/{patientId}/history [get]
func (h *PatientHandler) GetHistory(c echo.Context) error {
	patientUUID := strings.TrimSpace(c.Param("patientId"))
	if patientUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient id is required")
	}

	history, err := h.service.History(c.Request().Context(), patientUUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, history)
}

// Update updates a patient's demographic profile.
//
//	@Summary      Update patient
//	@Description  Partially updates a patient record. Only the fields supplied in the body are changed.
//	@Tags         Patients
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        patientId  path      string                 true  "Patient UUID"
//	@Param        body       body      UpdatePatientRequest   true  "Fields to update"
//	@Success      200        {array}   domain.PatientProfile  "Single-element array with the updated patient profile"
//	@Failure      400        {object}  ErrorResponse          "Validation error — invalid age_unit etc."
//	@Failure      401        {object}  ErrorResponse          "Missing or invalid JWT"
//	@Failure      500        {object}  ErrorResponse          "Unexpected server error"
//	@Router       /api/patients/{patientId} [patch]
func (h *PatientHandler) Update(c echo.Context) error {
	patientUUID := strings.TrimSpace(c.Param("patientId"))
	if patientUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient id is required")
	}

	var req UpdatePatientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if req.Age < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "age must be greater than or equal to 0")
	}

	ageUnit, err := normalizeAgeUnit(req.AgeUnit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	profile, err := h.service.UpdateProfile(c.Request().Context(), patientUUID, domain.PatientProfileUpdate{
		Prefix:      strings.TrimSpace(req.Prefix),
		FirstName:   strings.TrimSpace(req.FirstName),
		Gender:      strings.TrimSpace(req.Gender),
		Age:         req.Age,
		AgeUnit:     ageUnit,
		PhoneNo:     strings.TrimSpace(req.PhoneNo),
		PatientType: strings.TrimSpace(req.PatientType),
		Status:      strings.TrimSpace(req.Status),
		UpdatedBy:   strings.TrimSpace(req.UpdatedBy),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.PatientProfile{profile})
}

func (h *PatientHandler) SearchByPatientNo(c echo.Context) error {
	patientNo := strings.TrimSpace(c.Param("patientNo"))
	if patientNo == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient no is required")
	}
	if _, err := strconv.ParseInt(patientNo, 10, 64); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid patient no")
	}

	results, err := h.service.Search(c.Request().Context(), patientNo)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, results)
}

func normalizeAgeUnit(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return "Yrs", nil
	}

	switch v {
	case "yrs", "yr", "year", "years":
		return "Yrs", nil
	case "mon", "month", "months":
		return "Mon", nil
	case "days", "day":
		return "Days", nil
	default:
		return "", errors.New("age_unit must be one of: Yrs, Mon, Days")
	}
}

