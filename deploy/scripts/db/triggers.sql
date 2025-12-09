CREATE TRIGGER set_group_code AFTER INSERT ON dbtable_schema.groups FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.make_generic_code('code', 8);
CREATE TRIGGER set_group_code AFTER INSERT ON dbtable_schema.payments FOR EACH ROW EXECUTE FUNCTION dbfunc_schema.make_generic_code('code', 12);
