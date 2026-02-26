package domain

import "time"

type User struct {
	ID           string
	UserUUID     string
	GroupID      *int64
	GroupCode    string
	GroupName    string
	LoginID      string
	Username     string
	DisplayName  string
	Email        string
	Phone        string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
	LastLogin    *time.Time
}

type Patient struct {
	ID          int64
	PatientUUID string
	PatientNo   int64
	Prefix      string
	FirstName   string
	Gender      string
	Age         int
	AgeUnit     string
	Phone       string
	OPIPNo      string
	PatientType string
	CreatedBy   string
	Status      string
	CreatedAt   time.Time
}

type PatientSearchResult struct {
	PatientID   string `json:"patient_id"`
	FullName    string `json:"full_name"`
	PhoneNo     string `json:"phone_no"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	PatientNo   int64  `json:"patient_no"`
	PatientType string `json:"patient_type"`
	Status      string `json:"status"`
}

type PatientHistoryItem struct {
	BillDate    string  `json:"bill_date"`
	BillNo      int64   `json:"bill_no"`
	ServiceName string  `json:"service_name"`
	Status      string  `json:"status"`
	Rate        float64 `json:"rate"`
}

type PatientProfileUpdate struct {
	Prefix      string
	FirstName   string
	Gender      string
	Age         int
	AgeUnit     string
	PhoneNo     string
	PatientType string
	Status      string
	UpdatedBy   string
}

type PatientProfile struct {
	PatientID   string `json:"patient_id"`
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

type LabBill struct {
	ID             int64
	BillUUID       string
	BillNo         int64
	PatientID      int64
	PatientUUID    string
	VisitID        *int64
	DoctorID       *int64
	ReferredBy     string
	HospitalName   string
	TotalAmount    float64
	Discount       float64
	Tax            float64
	NetAmount      float64
	ReceivedAmount float64
	BalanceAmount  float64
	PaymentStatus  string
	Status         string
	PaymentMode    string
	CreatedAt      time.Time
}

type BillService struct {
	ServiceID   string
	ServiceName string
	Rate        float64
}

type LabService struct {
	ServiceID   int64   `json:"service_id"`
	ServiceName string  `json:"service_name"`
	Rate        float64 `json:"rate"`
	Department  string  `json:"department"`
	Status      string  `json:"status"`
}

type BillPaymentUpdate struct {
	BillID        string  `json:"bill_id"`
	BillNo        int64   `json:"bill_no"`
	NetBilledAmt  float64 `json:"net_billed_amt"`
	ReceivedAmt   float64 `json:"received_amt"`
	BalanceAmt    float64 `json:"balance_amt"`
	PaymentMode   string  `json:"payment_mode"`
	PaymentStatus string  `json:"payment_status"`
	UpdatedAt     string  `json:"updated_at"`
}

type LabOrder struct {
	ID        int64
	OrderUUID string
	BillID    int64
	PatientID int64
	VisitID   *int64
	Status    string
	CreatedAt time.Time
}

type ResultVerificationParam struct {
	ParamID     string `json:"param_id"`
	ParamName   string `json:"param_name"`
	ResultValue string `json:"result_value"`
	IsAbnormal  bool   `json:"is_abnormal"`
}

type SampleCollectionResponse struct {
	BillID             string `json:"bill_id"`
	SampleNo           string `json:"sample_no"`
	CollectionStatus   string `json:"collection_status"`
	CollectedBy        string `json:"collected_by"`
	WorksheetNo        string `json:"worksheet_no"`
	CollectionDateTime string `json:"collection_datetime"`
}

type ResultVerificationResponse struct {
	BillID               string `json:"bill_id"`
	VerificationStatus   string `json:"verification_status"`
	VerifiedBy           string `json:"verified_by"`
	VerificationDateTime string `json:"verification_datetime"`
	ResultCount          int    `json:"result_count"`
}

type ResultCertificationResponse struct {
	BillID               string `json:"bill_id"`
	CertificationStatus  string `json:"certification_status"`
	CertifiedBy          string `json:"certified_by"`
	CertificationRemarks string `json:"certification_remarks"`
	DispatchReady        bool   `json:"dispatch_ready"`
}

type LabReportResult struct {
	ParamName   string `json:"param_name"`
	ResultValue string `json:"result_value"`
	Reference   string `json:"reference"`
	Flag        string `json:"flag"`
}

type LabReportResponse struct {
	BillID         string            `json:"bill_id"`
	PatientID      string            `json:"patient_id"`
	PatientName    string            `json:"patient_name"`
	VerificationBy string            `json:"verification_by"`
	CertifiedBy    string            `json:"certified_by"`
	ReportStatus   string            `json:"report_status"`
	Results        []LabReportResult `json:"results"`
}
