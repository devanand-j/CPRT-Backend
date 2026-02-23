package domain

import "time"

type Role string

const (
	RoleSuperAdmin   Role = "SUPER_ADMIN"
	RoleAdmin        Role = "ADMIN"
	RoleDoctor       Role = "DOCTOR"
	RoleReceptionist Role = "RECEPTIONIST"
	RoleTechnician   Role = "TECHNICIAN"
)

type User struct {
	ID           int64
	UserUUID     string
	Username     string
	Email        string
	Phone        string
	PasswordHash string
	Role         Role
	Status       string
	CreatedAt    time.Time
}

type Patient struct {
	ID          int64
	PatientUUID string
	MRN         string
	Name        string
	Gender      string
	DOB         time.Time
	Phone       string
	Email       string
	Address     string
	CreatedAt   time.Time
}

type LabBill struct {
	ID          int64
	BillUUID    string
	PatientID   int64
	VisitID     *int64
	DoctorID    *int64
	TotalAmount float64
	Discount    float64
	Tax         float64
	NetAmount   float64
	Status      string
	PaymentMode string
	CreatedAt   time.Time
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
