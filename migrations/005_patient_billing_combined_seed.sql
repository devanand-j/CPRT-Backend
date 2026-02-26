-- Combined migration and seed data for Patient Master and Billing modules
-- Run this in pgAdmin Query Tool after 001_core_lis_schema.sql

BEGIN;

-- Patient Master fields
ALTER TABLE patients
    ADD COLUMN IF NOT EXISTS patient_type TEXT,
    ADD COLUMN IF NOT EXISTS created_by TEXT,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'Active';

-- Billing Financials fields
ALTER TABLE lab_bills
    ADD COLUMN IF NOT EXISTS bill_no BIGSERIAL UNIQUE NOT NULL,
    ADD COLUMN IF NOT EXISTS referred_by TEXT,
    ADD COLUMN IF NOT EXISTS hospital_name TEXT,
    ADD COLUMN IF NOT EXISTS received_amount NUMERIC(12,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS balance_amount NUMERIC(12,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payment_status TEXT DEFAULT 'Pending';

-- Seed 2 dummy patients
INSERT INTO patients (
    patient_uuid,
    prefix,
    first_name,
    gender,
    age,
    age_unit,
    phone,
    op_ip_no,
    patient_type,
    created_by,
    status,
    created_at
)
SELECT
    gen_random_uuid(),
    v.prefix,
    v.first_name,
    v.gender,
    v.age,
    v.age_unit,
    v.phone,
    v.op_ip_no,
    v.patient_type,
    v.created_by,
    v.status,
    NOW()
FROM (
    VALUES
        ('Mr.', 'John Doe', 'Male', 25, 'Yrs', '9876543210', 'OP-1002', 'Out Patients', 'USR-10827', 'Active'),
        ('Ms.', 'Priya Sharma', 'Female', 31, 'Yrs', '9123456780', 'OP-1003', 'Out Patients', 'USR-10827', 'Active')
) AS v(prefix, first_name, gender, age, age_unit, phone, op_ip_no, patient_type, created_by, status)
WHERE NOT EXISTS (
    SELECT 1
    FROM patients p
    WHERE p.op_ip_no = v.op_ip_no
);

COMMIT;

-- Verify data
SELECT patient_uuid AS "PAT-UUID-0001 (use for billing)", patient_no, prefix, first_name, phone, patient_type, status
FROM patients
ORDER BY patient_no DESC
LIMIT 5;
