DROP INDEX IF EXISTS uq_lab_results_bill_param;

CREATE UNIQUE INDEX IF NOT EXISTS uq_lab_results_bill_param_full
	ON lab_results (bill_id, param_id);
