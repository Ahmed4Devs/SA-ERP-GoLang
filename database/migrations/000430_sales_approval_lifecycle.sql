-- +goose Up
ALTER TABLE sales.invoices DROP CONSTRAINT invoices_status_check;
ALTER TABLE sales.invoices ADD CONSTRAINT invoices_status_check CHECK(status IN ('DRAFT','SUBMITTED','APPROVED','POSTED','CANCELLED'));
ALTER TABLE sales.invoices
 ADD COLUMN approval_instance_id UUID REFERENCES workflow.instances(id),
 ADD COLUMN approved_by UUID REFERENCES iam.users(id),
 ADD COLUMN approved_at TIMESTAMPTZ;

-- +goose StatementBegin
CREATE FUNCTION sales.guard_invoice_approval() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='DELETE' THEN
  IF OLD.status<>'DRAFT' THEN RAISE EXCEPTION 'only draft invoices may be deleted'; END IF;
  RETURN OLD;
 END IF;
 IF OLD.status IN ('SUBMITTED','APPROVED','POSTED') AND
   (to_jsonb(NEW)-ARRAY['status','journal_entry_id','open_item_id','qr_code_data','posted_at','updated_at','approved_by','approved_at','approval_instance_id']) IS DISTINCT FROM
   (to_jsonb(OLD)-ARRAY['status','journal_entry_id','open_item_id','qr_code_data','posted_at','updated_at','approved_by','approved_at','approval_instance_id']) THEN
  RAISE EXCEPTION 'submitted and posted invoice financial snapshots are immutable';
 END IF;
 IF OLD.status='POSTED' AND
   (to_jsonb(NEW)-ARRAY['qr_code_data','updated_at']) IS DISTINCT FROM (to_jsonb(OLD)-ARRAY['qr_code_data','updated_at']) THEN
  RAISE EXCEPTION 'posted invoice history is immutable; issue a credit note';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER invoice_approval_immutability BEFORE UPDATE OR DELETE ON sales.invoices FOR EACH ROW EXECUTE FUNCTION sales.guard_invoice_approval();
CREATE FUNCTION sales.guard_invoice_line_approval() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE state TEXT; parent UUID;
BEGIN
 IF TG_OP='INSERT' THEN parent:=NEW.invoice_id; ELSE parent:=OLD.invoice_id; END IF;
 SELECT status INTO state FROM sales.invoices WHERE id=parent FOR UPDATE;
 IF state IS DISTINCT FROM 'DRAFT' THEN RAISE EXCEPTION 'only draft invoice lines may change'; END IF;
 IF TG_OP='UPDATE' AND NEW.invoice_id IS DISTINCT FROM OLD.invoice_id THEN
  SELECT status INTO state FROM sales.invoices WHERE id=NEW.invoice_id FOR UPDATE;
  IF state IS DISTINCT FROM 'DRAFT' THEN RAISE EXCEPTION 'only draft invoice lines may change'; END IF;
 END IF;
 IF TG_OP='DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER invoice_line_approval_immutability BEFORE INSERT OR UPDATE OR DELETE ON sales.invoice_lines FOR EACH ROW EXECUTE FUNCTION sales.guard_invoice_line_approval();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER invoice_line_approval_immutability ON sales.invoice_lines;
DROP FUNCTION sales.guard_invoice_line_approval();
DROP TRIGGER invoice_approval_immutability ON sales.invoices;
DROP FUNCTION sales.guard_invoice_approval();
ALTER TABLE sales.invoices DROP COLUMN approved_at, DROP COLUMN approved_by, DROP COLUMN approval_instance_id;
ALTER TABLE sales.invoices DROP CONSTRAINT invoices_status_check;
ALTER TABLE sales.invoices ADD CONSTRAINT invoices_status_check CHECK(status IN ('DRAFT','APPROVED','POSTED','CANCELLED'));
