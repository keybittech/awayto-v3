-- code generation
CREATE TRIGGER set_group_code AFTER INSERT ON dbtable_schema.groups FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.make_generic_code('code', 8);
CREATE TRIGGER set_group_code AFTER INSERT ON dbtable_schema.seat_payments FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.make_generic_code('code', 12);

-- seat handling
CREATE TRIGGER trg_seat_payments_balance AFTER INSERT OR UPDATE ON dbtable_schema.seat_payments FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.trg_handle_seat_payment();
