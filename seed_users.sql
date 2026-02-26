-- Run this after migrations/001_core_lis_schema.sql
-- Password for all users: password123 (bcrypt)

INSERT INTO account_groups (group_code, group_name, status) VALUES
('SUPER_ADMIN', 'Super Admin', 'ACTIVE'),
('ADMIN', 'Admin', 'ACTIVE'),
('DOCTOR', 'Doctor', 'ACTIVE'),
('RECEPTIONIST', 'Receptionist', 'ACTIVE'),
('TECHNICIAN', 'Technician', 'ACTIVE')
ON CONFLICT (group_code) DO NOTHING;

INSERT INTO users (user_uuid, username, login_id, user_name, email, password_hash, group_id, status, created_at, last_login) VALUES
(gen_random_uuid(), 'superadmin', 'superadmin', 'Super Admin', 'superadmin@cprt.com', '$2a$10$QllwYX0JhkXaKlK2DEsfq.dpaDBz.sWPrRm0.6isKFqVkCwbop4Ym', (SELECT id FROM account_groups WHERE group_code = 'SUPER_ADMIN'), 'ACTIVE', NOW(), NOW()),
(gen_random_uuid(), 'admin', 'admin', 'Admin User', 'admin@cprt.com', '$2a$10$QllwYX0JhkXaKlK2DEsfq.dpaDBz.sWPrRm0.6isKFqVkCwbop4Ym', (SELECT id FROM account_groups WHERE group_code = 'ADMIN'), 'ACTIVE', NOW(), NOW()),
(gen_random_uuid(), 'doctor1', 'doctor1', 'Doctor One', 'doctor@cprt.com', '$2a$10$QllwYX0JhkXaKlK2DEsfq.dpaDBz.sWPrRm0.6isKFqVkCwbop4Ym', (SELECT id FROM account_groups WHERE group_code = 'DOCTOR'), 'ACTIVE', NOW(), NOW()),
(gen_random_uuid(), 'receptionist1', 'receptionist1', 'Reception One', 'reception@cprt.com', '$2a$10$QllwYX0JhkXaKlK2DEsfq.dpaDBz.sWPrRm0.6isKFqVkCwbop4Ym', (SELECT id FROM account_groups WHERE group_code = 'RECEPTIONIST'), 'ACTIVE', NOW(), NOW()),
(gen_random_uuid(), 'tech1', 'tech1', 'Tech One', 'tech@cprt.com', '$2a$10$QllwYX0JhkXaKlK2DEsfq.dpaDBz.sWPrRm0.6isKFqVkCwbop4Ym', (SELECT id FROM account_groups WHERE group_code = 'TECHNICIAN'), 'ACTIVE', NOW(), NOW())
ON CONFLICT (username) DO UPDATE
SET
	login_id = EXCLUDED.login_id,
	user_name = EXCLUDED.user_name,
	email = EXCLUDED.email,
	password_hash = EXCLUDED.password_hash,
	group_id = EXCLUDED.group_id,
	status = EXCLUDED.status;

INSERT INTO user_account_groups (user_id, group_id)
SELECT u.id, g.id
FROM users u
JOIN account_groups g
	ON (u.username = 'superadmin' AND g.group_code = 'SUPER_ADMIN')
	OR (u.username = 'admin' AND g.group_code = 'ADMIN')
	OR (u.username = 'doctor1' AND g.group_code = 'DOCTOR')
	OR (u.username = 'receptionist1' AND g.group_code = 'RECEPTIONIST')
	OR (u.username = 'tech1' AND g.group_code = 'TECHNICIAN')
ON CONFLICT (user_id, group_id) DO NOTHING;

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
	SELECT 1 FROM patients p WHERE p.op_ip_no = v.op_ip_no
);
