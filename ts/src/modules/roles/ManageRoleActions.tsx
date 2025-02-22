import React, { useCallback, useEffect, useMemo, useState } from 'react';

import Card from '@mui/material/Card';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Grid from '@mui/material/Grid';
import CardActionArea from '@mui/material/CardActionArea';
import Checkbox from '@mui/material/Checkbox';

import { DataGrid, GridColDef } from '@mui/x-data-grid';

import { useGrid, siteApi, useUtil, SiteRoles, SiteRoleDetails } from 'awayto/hooks';
import { Tooltip } from '@mui/material';

export function ManageRoleActions(_: IComponent): React.JSX.Element {

  const { setSnack } = useUtil();
  const [patchAssignments] = siteApi.useGroupServicePatchGroupAssignmentsMutation();

  const { data: availableGroupAssignmentsRequest } = siteApi.useGroupServiceGetGroupAssignmentsQuery();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  const [assignments, setAssignments] = useState(availableGroupAssignmentsRequest?.assignments || {});

  useEffect(() => availableGroupAssignmentsRequest?.assignments && setAssignments(availableGroupAssignmentsRequest?.assignments), [availableGroupAssignmentsRequest?.assignments]);

  const groupsValues = useMemo(() => Object.values(profileRequest?.userProfile?.groups || {}), [profileRequest?.userProfile]);

  const handleCheck = useCallback((subgroup: string, action: string, add: boolean) => {
    const newAssignments = { ...assignments };
    newAssignments[subgroup] = { actions: add ? [...(newAssignments[subgroup]?.actions || []), { name: action }] : newAssignments[subgroup]?.actions?.filter(a => a.name !== action) };
    setAssignments(newAssignments);
  }, [assignments]);

  const handleSubmit = useCallback(() => {
    const patchedAssignments = { ...assignments };
    const group = groupsValues.find(g => g.active);
    if (group) {
      delete patchedAssignments[`/${group.name}/Admin`]
      patchAssignments({ patchGroupAssignmentsRequest: { assignments: patchedAssignments } }).unwrap().then(() => {
        setSnack({ snackType: 'success', snackOn: 'Assignments can be updated again in 1 minute.' });
      }).catch(console.error);
    }
  }, [assignments]);

  const columns = useMemo(() => {
    if (groupsValues.length) {
      const group = groupsValues.find(g => g.active);
      if (group?.roles && Object.keys(group.roles).length) {

        const cols: GridColDef<{ id: string, name: string, description: string }>[] = [{
          width: 200,
          field: 'id',
          headerName: '',
          cellClassName: 'vertical-parent',
          renderCell: ({ row }) => <Tooltip placement="right" title={row.description}>
            <Typography>{row.name}</Typography>
          </Tooltip>
        }];

        for (const roleId in group.roles) {
          const { name } = group.roles[roleId];
          if (!name) continue;
          const subgroup = `/${group.name}/${name}`;
          cols.push({
            flex: 1,
            minWidth: 75,
            headerName: name,
            field: name,
            renderCell: ({ row }) => {
              if (!assignments[subgroup]) return <></>;
              return <Checkbox
                disabled={name.toLowerCase() == "admin"}
                checked={assignments[subgroup].actions?.some(a => a.name === row.id) ?? false}
                onChange={e => handleCheck(subgroup, row.id, e.target.checked)}
              />
            }
          });
        }

        return cols;
      }
    }
    return [];
  }, [groupsValues, assignments]);

  const options = useMemo(() => {
    return Object.values(SiteRoles)
      .filter(r => ![SiteRoles.UNRESTRICTED, SiteRoles.APP_ROLE_CALL, SiteRoles.APP_GROUP_ADMIN].includes(r))
      .map(r => {
        return { id: r, name: SiteRoleDetails[r].name, description: SiteRoleDetails[r].description }
      });
  }, []);

  const roleActionGridProps = useGrid({
    rows: options,
    columns,
    noPagination: true
  });

  return <>

    <Grid container>
      <Grid mb={2} size={12}>
        <Card>
          <CardActionArea onClick={handleSubmit}>
            <Grid container direction="row" justifyContent="space-between">
              <Grid>
                <Box m={2}>
                  <Typography color="secondary" variant="button">Update Assignments </Typography>
                </Box>
              </Grid>
              <Grid>
                <Box m={2}>
                  <Typography color="GrayText" variant="button">Changes will persist within 1 minute</Typography>
                </Box>
              </Grid>
            </Grid>
          </CardActionArea>
        </Card>
      </Grid>
    </Grid>

    <DataGrid {...roleActionGridProps} />

  </>
}

export default ManageRoleActions;
