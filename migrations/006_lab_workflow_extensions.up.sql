ALTER TABLE lab_bills
	ADD COLUMN IF NOT EXISTS report_status TEXT DEFAULT 'Pending',
	ADD COLUMN IF NOT EXISTS certified_by_user TEXT,
	ADD COLUMN IF NOT EXISTS certification_remarks TEXT,
	ADD COLUMN IF NOT EXISTS certified_at TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS dispatch_ready BOOLEAN DEFAULT FALSE;

ALTER TABLE lab_results
	ADD COLUMN IF NOT EXISTS bill_id BIGINT REFERENCES lab_bills(id),
	ADD COLUMN IF NOT EXISTS param_id TEXT,
	ADD COLUMN IF NOT EXISTS param_name TEXT,
	ADD COLUMN IF NOT EXISTS verified_by_user TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uq_lab_results_bill_param
	ON lab_results (bill_id, param_id)
	WHERE bill_id IS NOT NULL AND param_id IS NOT NULL;
