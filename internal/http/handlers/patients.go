package handlers

import (
	"net/http"
	"strconv"
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
	MRN     string `json:"mrn"`
	Name    string `json:"name"`
	Gender  string `json:"gender"`
	DOB     string `json:"dob"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func (h *PatientHandler) Create(c echo.Context) error {
	var req createPatientRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid dob format")
	}

	patient, err := h.service.Create(c.Request().Context(), domain.Patient{
		MRN:     req.MRN,
		Name:    req.Name,
		Gender:  req.Gender,
		DOB:     dob,
		Phone:   req.Phone,
		Email:   req.Email,
		Address: req.Address,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, patient)
}

func (h *PatientHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	patient, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "patient not found")
	}

	return c.JSON(http.StatusOK, patient)
}

func (h *PatientHandler) Search(c echo.Context) error {
	mrn := c.QueryParam("mrn")
	phone := c.QueryParam("phone")

	patients, err := h.service.Search(c.Request().Context(), mrn, phone)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, patients)
}
