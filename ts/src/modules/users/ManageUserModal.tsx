import React, { useCallback, useState } from 'react';

import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import MenuItem from '@mui/material/MenuItem';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

import { IUserProfile, siteApi, targets } from 'awayto/hooks';

interface ManageUserModalProps extends IComponent {
  editRoleId?: string;
  editUser?: IUserProfile;
}

export function ManageUserModal({ editRoleId, editUser, closeModal }: ManageUserModalProps): React.JSX.Element {

  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const [patchGroupUser] = siteApi.useGroupUsersServicePatchGroupUserMutation();

  // const [password, setPassword] = useState('');
  const [roleId, setRoleId] = useState(editRoleId);
  const [profile, _] = useState({
    firstName: '',
    lastName: '',
    email: '',
    username: '',
    roleId: '',
    roleName: '',
    ...editUser
  } as IUserProfile);

  // const handlePassword = useCallback(({ target: { value } }: React.ChangeEvent<HTMLTextAreaElement>) => setPassword(value), [])
  // const handleProfile = useCallback(({ target: { name, value } }: React.ChangeEvent<HTMLTextAreaElement>) => setProfile({ ...profile, [name]: value }), [profile])

  const handleSubmit = useCallback(() => {
    async function go() {
      if (editUser?.id && groupRolesRequest?.groupRoles) {
        if (profile.sub && roleId) {
          await patchGroupUser({ patchGroupUserRequest: { userSub: profile.sub, roleId } }).unwrap();
          if (closeModal)
            closeModal();
        }
      }
    }
    void go();
    // async function submitUser() {
    //   let user = profile as IUserProfile;
    //   const { id, sub } = user;

    //   const groupRoleKeys = Object.keys(userGroupRoles);

    //   if (!groupRoleKeys.length)
    //     return act(SET_SNACK, { snackType: 'error', snackOn: 'Group roles must be assigned.' });

    //   user.groups = groupRoleKeys // { "g1": [...], "g2": [...] } => ["g1", "g2"];
    //     .reduce((memo, key) => {
    //       if (userGroupRoles[key].length) { // [...].length
    //         const group = { ...groups?.find(g => g.name == key) } as IGroup; // Get a copy from repository
    //         if (group.roles) {
    //           group.roles = group.roles.filter(r => userGroupRoles[key].includes(r.name)) // Filter roles
    //           memo.push(group);
    //         }
    //       }
    //       return memo;
    //     }, [] as IGroup[]);

    //   // User Update - 3 Scenarios
    //   // User-Originated - A user signed up in the wild, created a cognito account, has not done anything to 
    //   //                  generate an application account, and now an admin is generating one
    //   // Admin-Originated - A user being created by an admin in the manage area
    //   // Admin-Updated - A user being updated by an admin in the manage area

    //   // If there's already a sub, PUT/manage/users will update the sub in cognito;
    //   // else we'll POST/manage/users/sub for a new sub
    //   user = await api(sub ? PUT_MANAGE_USERS : POST_MANAGE_USERS_SUB, true, sub ? user : { ...user, password }) as IUserProfile;

    //   // Add user to application db if needed no user.id
    //   if (!id)
    //     user = await api(sub ? POST_MANAGE_USERS_APP_ACCT : POST_MANAGE_USERS, true, user) as IUserProfile;

    //   if (closeModal)
    //     closeModal();
    // }

    // void submitUser();
  }, [profile, roleId]);

  // const passwordGenerator = useCallback(() => {
  //   setPassword(passwordGen());
  // }, []);

  return <>
    <Card>
      <CardHeader
        title={`Manage ${profile.email}`}
        subheader={`${profile.firstName} ${profile.lastName}`}
      />
      <CardContent>
        <Grid container direction="row" spacing={2} justifyContent="space-evenly">
          <Grid size={12}>
            <Grid container direction="column" spacing={4} justifyContent="space-evenly" >

              <Grid>
                <TextField
                  {...targets(`manage user modal role selection`, `Role`, `edit the user's role`)}
                  select
                  fullWidth
                  value={roleId}
                  onChange={e => setRoleId(e.target.value)}
                >
                  {groupRolesRequest?.groupRoles?.map(gr =>
                    <MenuItem key={`${gr.role?.id}_user_profile_role_select`} value={gr.role?.id}>{gr.role?.name}</MenuItem>
                  )}
                </TextField>
              </Grid>

              {/* !editUser && (
                <Grid xs>
                  <Grid container direction="column" justifyContent="space-evenly" spacing={4}>
                    <Grid size={12}>
                      <Typography variant="h6">Account</Typography>
                    </Grid>
                    <Grid size={12}>
                      <TextField fullWidth id="username" label="Username" value={profile.username} name="username" onChange={handleProfile} />
                    </Grid>
                    <Grid size={12}>
                      <FormControl fullWidth>
                        <InputLabel htmlFor="password">Password</InputLabel>
                        <Input type="text" id="password" aria-describedby="password" value={password} onChange={handlePassword}
                          endAdornment={
                            <InputAdornment position="end">
                              <Button onClick={passwordGenerator} style={{ backgroundColor: 'transparent' }}>Generate</Button>
                            </InputAdornment>
                          }
                        />
                        <FormHelperText>Password must be at least 8 characters and contain 1 uppercase, lowercase, number, and special (e.g. @^$!*) character. The user must change this passowrd upon logging in for the first time.</FormHelperText>
                      </FormControl>
                    </Grid>
                  </Grid>
                </Grid>
              ) */}
            </Grid>
          </Grid>
        </Grid>
      </CardContent>
      <CardActions>

        <Grid container justifyContent="space-between">
          <Button
            {...targets(`manage users modal close`, `close the user management modal`)}
            onClick={closeModal}
          >Cancel</Button>
          <Button
            {...targets(`mange users modal submit`, `submit changes to the current user`)}
            onClick={handleSubmit}
          >{profile.sub ? 'update' : 'create'}</Button>
        </Grid>
      </CardActions>
    </Card>
  </>
}

export default ManageUserModal;
