CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS roles (
	id BIGSERIAL PRIMARY KEY,
	role_code TEXT UNIQUE NOT NULL,
	role_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	user_uuid UUID UNIQUE NOT NULL,
	username TEXT UNIQUE NOT NULL,
	email TEXT,
	phone TEXT,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('SUPER_ADMIN', 'ADMIN', 'DOCTOR', 'RECEPTIONIST', 'TECHNICIAN')),
	status TEXT NOT NULL DEFAULT 'ACTIVE',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_roles (
	user_id BIGINT REFERENCES users(id),
	role_id BIGINT REFERENCES roles(id),
	assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS account_groups (
	id BIGSERIAL PRIMARY KEY,
	group_code TEXT UNIQUE NOT NULL,
	group_name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'ACTIVE'
);

CREATE TABLE IF NOT EXISTS user_account_groups (
	user_id BIGINT REFERENCES users(id),
	group_id BIGINT REFERENCES account_groups(id),
	PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS hospitals (
	id BIGSERIAL PRIMARY KEY,
	hospital_code TEXT UNIQUE NOT NULL,
	hospital_name TEXT NOT NULL,
	address TEXT
);

CREATE TABLE IF NOT EXISTS wards (
	id BIGSERIAL PRIMARY KEY,
	ward_code TEXT UNIQUE NOT NULL,
	ward_name TEXT NOT NULL,
	hospital_id BIGINT REFERENCES hospitals(id)
);

CREATE TABLE IF NOT EXISTS patients (
	id BIGSERIAL PRIMARY KEY,
	patient_uuid UUID UNIQUE NOT NULL,
	mrn TEXT UNIQUE,
	name TEXT NOT NULL,
	gender TEXT,
	dob DATE,
	phone TEXT,
	email TEXT,
	address TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS patient_demographics (
	patient_id BIGINT PRIMARY KEY REFERENCES patients(id),
	marital_status TEXT,
	blood_group TEXT,
	nationality TEXT,
	occupation TEXT,
	emergency_contact TEXT
);

CREATE TABLE IF NOT EXISTS patient_visits (
	id BIGSERIAL PRIMARY KEY,
	visit_uuid UUID UNIQUE NOT NULL,
	patient_id BIGINT REFERENCES patients(id),
	visit_type TEXT NOT NULL,
	visit_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	doctor_id BIGINT REFERENCES users(id),
	ward_id BIGINT REFERENCES wards(id),
	status TEXT NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE IF NOT EXISTS lab_services (
	id BIGSERIAL PRIMARY KEY,
	service_code TEXT UNIQUE NOT NULL,
	service_name TEXT NOT NULL,
	specimen_type TEXT,
	department TEXT,
	base_price NUMERIC(12,2) NOT NULL,
	tat_minutes INT,
	status TEXT NOT NULL DEFAULT 'ACTIVE'
);

CREATE TABLE IF NOT EXISTS price_profiles (
	id BIGSERIAL PRIMARY KEY,
	profile_code TEXT UNIQUE NOT NULL,
	profile_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS service_prices (
	service_id BIGINT REFERENCES lab_services(id),
	profile_id BIGINT REFERENCES price_profiles(id),
	price NUMERIC(12,2) NOT NULL,
	valid_from DATE,
	valid_to DATE,
	PRIMARY KEY (service_id, profile_id)
);

CREATE TABLE IF NOT EXISTS lab_bills (
	id BIGSERIAL PRIMARY KEY,
	bill_uuid UUID UNIQUE NOT NULL,
	patient_id BIGINT REFERENCES patients(id),
	visit_id BIGINT REFERENCES patient_visits(id),
	doctor_id BIGINT REFERENCES users(id),
	total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
	discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
	tax_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
	net_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'DRAFT',
	payment_mode TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS lab_bill_items (
	id BIGSERIAL PRIMARY KEY,
	bill_id BIGINT REFERENCES lab_bills(id),
	service_id BIGINT REFERENCES lab_services(id),
	qty INT NOT NULL DEFAULT 1,
	unit_price NUMERIC(12,2) NOT NULL,
	discount NUMERIC(12,2) NOT NULL DEFAULT 0,
	tax NUMERIC(12,2) NOT NULL DEFAULT 0,
	line_total NUMERIC(12,2) NOT NULL,
	status TEXT NOT NULL DEFAULT 'PENDING'
);

CREATE TABLE IF NOT EXISTS lab_orders (
	id BIGSERIAL PRIMARY KEY,
	order_uuid UUID UNIQUE NOT NULL,
	bill_id BIGINT REFERENCES lab_bills(id),
	patient_id BIGINT REFERENCES patients(id),
	visit_id BIGINT REFERENCES patient_visits(id),
	order_status TEXT NOT NULL DEFAULT 'PENDING',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
	id BIGSERIAL PRIMARY KEY,
	order_id BIGINT REFERENCES lab_orders(id),
	service_id BIGINT REFERENCES lab_services(id),
	status TEXT NOT NULL DEFAULT 'PENDING'
);

CREATE TABLE IF NOT EXISTS samples (
	id BIGSERIAL PRIMARY KEY,
	sample_uuid UUID UNIQUE NOT NULL,
	order_id BIGINT REFERENCES lab_orders(id),
	specimen_type TEXT,
	barcode TEXT UNIQUE,
	collected_by BIGINT REFERENCES users(id),
	collected_at TIMESTAMPTZ,
	status TEXT NOT NULL DEFAULT 'PENDING'
);

CREATE TABLE IF NOT EXISTS worksheets (
	id BIGSERIAL PRIMARY KEY,
	worksheet_uuid UUID UNIQUE NOT NULL,
	created_by BIGINT REFERENCES users(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	status TEXT NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE IF NOT EXISTS worksheet_samples (
	worksheet_id BIGINT REFERENCES worksheets(id),
	sample_id BIGINT REFERENCES samples(id),
	PRIMARY KEY (worksheet_id, sample_id)
);

CREATE TABLE IF NOT EXISTS lab_results (
	id BIGSERIAL PRIMARY KEY,
	order_item_id BIGINT REFERENCES order_items(id),
	result_value TEXT,
	unit TEXT,
	reference_range TEXT,
	abnormal_flag TEXT,
	result_status TEXT NOT NULL DEFAULT 'ENTERED',
	verified_by BIGINT REFERENCES users(id),
	verified_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS audit_logs (
	id BIGSERIAL PRIMARY KEY,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	action TEXT NOT NULL,
	actor_id BIGINT REFERENCES users(id),
	payload_json JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patients_phone ON patients(phone);
CREATE INDEX IF NOT EXISTS idx_patients_mrn ON patients(mrn);
CREATE INDEX IF NOT EXISTS idx_lab_orders_status ON lab_orders(order_status);
CREATE INDEX IF NOT EXISTS idx_lab_bills_status ON lab_bills(status);
