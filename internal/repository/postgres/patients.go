package postgres

import (
	"context"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PatientRepository struct {
	pool *pgxpool.Pool
}

func NewPatientRepository(pool *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{pool: pool}
}

func (r *PatientRepository) Create(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	const query = `
		INSERT INTO patients (patient_uuid, mrn, name, gender, dob, phone, email, address)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		RETURNING id, patient_uuid, created_at
	`
	row := r.pool.QueryRow(ctx, query, patient.MRN, patient.Name, patient.Gender, patient.DOB, patient.Phone, patient.Email, patient.Address)
	if err := row.Scan(&patient.ID, &patient.PatientUUID, &patient.CreatedAt); err != nil {
		return domain.Patient{}, err
	}
	return patient, nil
}

func (r *PatientRepository) GetByID(ctx context.Context, id int64) (domain.Patient, error) {
	const query = `
		SELECT id, patient_uuid, mrn, name, gender, dob, phone, email, address, created_at
		FROM patients
		WHERE id = $1
	`
	var patient domain.Patient
	row := r.pool.QueryRow(ctx, query, id)
	if err := row.Scan(&patient.ID, &patient.PatientUUID, &patient.MRN, &patient.Name, &patient.Gender, &patient.DOB, &patient.Phone, &patient.Email, &patient.Address, &patient.CreatedAt); err != nil {
		return domain.Patient{}, err
	}
	return patient, nil
}

func (r *PatientRepository) Search(ctx context.Context, mrn, phone string) ([]domain.Patient, error) {
	const query = `
		SELECT id, patient_uuid, mrn, name, gender, dob, phone, email, address, created_at
		FROM patients
		WHERE ($1 = '' OR mrn = $1)
		  AND ($2 = '' OR phone = $2)
		ORDER BY created_at DESC
		LIMIT 100
	`
	rows, err := r.pool.Query(ctx, query, mrn, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patients := []domain.Patient{}
	for rows.Next() {
		var patient domain.Patient
		if err := rows.Scan(&patient.ID, &patient.PatientUUID, &patient.MRN, &patient.Name, &patient.Gender, &patient.DOB, &patient.Phone, &patient.Email, &patient.Address, &patient.CreatedAt); err != nil {
			return nil, err
		}
		patients = append(patients, patient)
	}
	return patients, nil
}
