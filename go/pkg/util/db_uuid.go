package util

import "github.com/jackc/pgx/v5/pgtype"

func RegisterUUID(tm *pgtype.Map) {
	tm.RegisterType(&pgtype.Type{
		Name:  "uuid",
		OID:   pgtype.UUIDOID,
		Codec: &pgtype.TextCodec{},
	})
	uuidType, _ := tm.TypeForOID(pgtype.UUIDOID)
	tm.RegisterType(&pgtype.Type{
		Name:  "_uuid",
		OID:   pgtype.UUIDArrayOID,
		Codec: &pgtype.ArrayCodec{ElementType: uuidType},
	})
}
