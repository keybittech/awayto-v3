package api

import (
	"context"
	"log"
	"sync"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

const getGroupDetailsQuery = `
	SELECT id, name, external_id, sub, ai
	FROM dbtable_schema.groups
	WHERE enabled = true
`

func (a *API) InitGroups() {
	workerDbSession := &clients.DbSession{
		Pool: a.Handlers.Database.DatabaseClient.Pool,
		UserSession: &types.UserSession{
			UserSub: "worker",
		},
	}

	ctx := context.Background()

	rows, done, err := workerDbSession.SessionBatchQuery(ctx, getGroupDetailsQuery)
	if err != nil {
		log.Fatal(util.ErrCheck(err))
	}
	defer done()

	var wg sync.WaitGroup
	for rows.Next() {
		dbGroup := &types.IGroup{}

		err := rows.Scan(&dbGroup.Id, &dbGroup.Name, &dbGroup.ExternalId, &dbGroup.CreatedSub, &dbGroup.Ai)
		if err != nil {
			log.Fatal(util.ErrCheck(err))
		}

		wg.Add(1)
		go func(g *types.IGroup) {
			defer wg.Done()
			ctx := context.Background()

			kcGroup, err := a.Handlers.Keycloak.GetGroup(ctx, "worker", g.ExternalId)
			if err != nil {
				util.DebugLog.Printf("%s", util.ErrCheck(err))
				return
			}

			a.Handlers.Cache.SetCachedGroup(kcGroup.Path, g.Id, g.ExternalId, g.CreatedSub, g.Name, g.Ai)

			kcSubGroups, err := a.Handlers.Keycloak.GetGroupSubGroups(ctx, "worker", g.ExternalId)
			if err != nil {
				util.DebugLog.Printf("%s", util.ErrCheck(err))
				return
			}

			for _, subGroup := range kcSubGroups {
				a.Handlers.Cache.SetCachedSubGroup(subGroup.Path, subGroup.Id, subGroup.Name, kcGroup.Path)
			}

			a.Handlers.Cache.SetGroupSessionVersion(g.Id)
		}(dbGroup)
	}

	wg.Wait()
}
