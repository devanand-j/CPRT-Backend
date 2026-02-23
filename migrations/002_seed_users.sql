-- Sample users with SHA256 hashed password "password123"
-- SHA256("password123") = ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f

INSERT INTO users (user_uuid, username, email, password_hash, role, status) VALUES
(gen_random_uuid(), 'superadmin', 'superadmin@cprt.com', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f', 'SUPER_ADMIN', 'ACTIVE'),
(gen_random_uuid(), 'admin', 'admin@cprt.com', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f', 'ADMIN', 'ACTIVE'),
(gen_random_uuid(), 'doctor1', 'doctor@cprt.com', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f', 'DOCTOR', 'ACTIVE'),
(gen_random_uuid(), 'receptionist1', 'reception@cprt.com', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f', 'RECEPTIONIST', 'ACTIVE'),
(gen_random_uuid(), 'tech1', 'tech@cprt.com', 'ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f', 'TECHNICIAN', 'ACTIVE')
ON CONFLICT (username) DO NOTHING;
