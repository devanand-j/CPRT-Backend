package handlers

import (
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

type createPatientRequest struct {
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	Gender      string `json:"gender"`
	Age         int    `json:"age"`
	AgeUnit     string `json:"age_unit"`
	PhoneNo     string `json:"phone_no"`
	Phone       string `json:"phone"`
	OPIPNo      string `json:"op_ip_no"`
	PatientType string `json:"patient_type"`
	CreatedBy   string `json:"created_by"`
	Status      string `json:"status"`
}

type patientRegisterResponse struct {
	PatientID        string    `json:"patient_id"`
	PatientNo        int64     `json:"patient_no"`
	FullName         string    `json:"full_name"`
	RegistrationDate string    `json:"registration_date"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

func (h *PatientHandler) Register(c echo.Context) error {
	var req createPatientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if strings.TrimSpace(req.FirstName) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "first_name is required")
	}
	if req.Age < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "age must be greater than or equal to 0")
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
		AgeUnit:     strings.TrimSpace(req.AgeUnit),
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

	resp := []patientRegisterResponse{{
		PatientID:        patient.PatientUUID,
		PatientNo:        patient.PatientNo,
		FullName:         fullName,
		RegistrationDate: patient.CreatedAt.Format("2006-01-02"),
		Status:           patient.Status,
		CreatedAt:        patient.CreatedAt,
	}}

	return c.JSON(http.StatusCreated, resp)
}

func (h *PatientHandler) Create(c echo.Context) error {
	return h.Register(c)
}

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

type updatePatientRequest struct {
	Prefix      string `json:"prefix"`
	FirstName   string `json:"first_name"`
	Gender      string `json:"gender"`
	Age         int    `json:"age"`
	AgeUnit     string `json:"age_unit"`
	PhoneNo     string `json:"phone_no"`
	PatientType string `json:"patient_type"`
	Status      string `json:"status"`
	UpdatedBy   string `json:"updated_by"`
}

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

func (h *PatientHandler) Update(c echo.Context) error {
	patientUUID := strings.TrimSpace(c.Param("patientId"))
	if patientUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient id is required")
	}

	var req updatePatientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if req.Age < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "age must be greater than or equal to 0")
	}

	profile, err := h.service.UpdateProfile(c.Request().Context(), patientUUID, domain.PatientProfileUpdate{
		Prefix:      strings.TrimSpace(req.Prefix),
		FirstName:   strings.TrimSpace(req.FirstName),
		Gender:      strings.TrimSpace(req.Gender),
		Age:         req.Age,
		AgeUnit:     strings.TrimSpace(req.AgeUnit),
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
