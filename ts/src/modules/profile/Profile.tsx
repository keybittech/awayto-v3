import React, { useState, useEffect } from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

import { siteApi, useStyles, useUtil, IUserProfile, PatchUserProfileRequest } from 'awayto/hooks';
import PickTheme from '../common/PickTheme';
import ManageGroups from '../groups/ManageGroups';

export function Profile(props: IComponent): React.JSX.Element {
  const classes = useStyles();

  const { setSnack } = useUtil();
  const [patchUserProfile] = siteApi.useUserProfileServicePatchUserProfileMutation();

  // const fileStore = useFileStore();

  const { data: profileRequest } = siteApi.useUserProfileServiceGetUserProfileDetailsQuery();

  // const [displayImage, setDisplayImage] = useState('');
  // const [file, setFile] = useState<IPreviewFile>();
  const [profile, setProfile] = useState({
    firstName: '',
    lastName: '',
    email: '',
    // image: ''
  } as Required<IUserProfile>);

  // const { getRootProps, getInputProps } = useDropzone({
  //   maxSize: 1000000,
  //   maxFiles: 1,
  //   accept: {
  //     'image/*': []
  //   },
  //   onDrop: (acceptedFiles: File[]) => {
  //     const acceptedFile = acceptedFiles.pop()
  //     if (acceptedFile) {
  //       setFile(acceptedFile);
  //       setDisplayImage(URL.createObjectURL(acceptedFile));
  //     }
  //   }
  // });

  // useEffect(() => {
  //   if (file?.preview) URL.revokeObjectURL(file.preview);
  // }, [file]);

  // useEffect(() => {
  //   async function go() {
  //     if (fileStore && profile.image) {
  //       setDisplayImage(await fileStore.get(profile.image));
  //     }
  //   }
  //   void go();
  // }, [fileStore, profile.image]);

  useEffect(() => {
    if (profileRequest?.userProfile) {
      setProfile({ ...profile, ...profileRequest.userProfile });
    }
  }, [profileRequest]);

  // const deleteFile = () => {
  //   setProfile({ ...profile, ...{ image: '' } });
  //   setDisplayImage('');
  // }

  const handleSubmit = () => {
    async function go() {
      // if (file) {
      //   profile.image = await fileStore?.put(file);
      // }

      const { firstName, lastName, email } = profile;

      patchUserProfile({ patchUserProfileRequest: { firstName, lastName, email } as PatchUserProfileRequest }).unwrap().then(() => {
        setSnack({ snackType: 'success', snackOn: 'Profile updated!' });
        // setFile(undefined);
      }).catch(console.error);
    }
    void go();
  }

  return <>
    <Grid container spacing={6}>
      <Grid size={{ sm: 12, md: 4 }}>
        <Grid container direction="column" spacing={2}>
          <Grid>
            <Typography variant="h6">Profile</Typography>
          </Grid>
          <Grid>
            <TextField fullWidth id="firstName" label="First Name" autoComplete="on" value={profile.firstName} name="firstName" onChange={e => setProfile({ ...profile, firstName: e.target.value })} />
          </Grid>
          <Grid>
            <TextField fullWidth id="lastName" label="Last Name" autoComplete="on" value={profile.lastName} name="lastName" onChange={e => setProfile({ ...profile, lastName: e.target.value })} />
          </Grid>
          <Grid>
            <TextField fullWidth id="email" label="Email" autoComplete="on" value={profile.email} name="email" onChange={e => setProfile({ ...profile, email: e.target.value })} />
          </Grid>
          {/* <Grid>
            <Typography variant="h6">Image</Typography>
          </Grid>
          <Grid>
            <CardActionArea style={{ padding: '12px' }}>
              {!displayImage ?
                <Grid {...getRootProps()} container alignItems="center" direction="column">
                  <input {...getInputProps()} />
                  <Grid>
                    <Avatar>
                      <PersonIcon />
                    </Avatar>
                  </Grid>
                  <Grid>
                    <Typography variant="subtitle1">Click or drag and drop to add a profile pic.</Typography>
                  </Grid>
                  <Grid>
                    <Typography variant="caption">Max size: 1MB</Typography>
                  </Grid>
                </Grid> :
                <Grid onClick={deleteFile} container alignItems="center" direction="column">
                  <Grid>
                    <Avatar src={displayImage} /> 
                  </Grid>
                  <Grid>
                    <Typography variant="h6" style={{ wordBreak: 'break-all' }}>{profileRequest?.userProfile?.image ? "Current profile image." : file ? `${file.name || ''} added.` : ''}</Typography>
                  </Grid>
                  <Grid>
                    <Typography variant="subtitle1">To remove, click here then submit.</Typography>
                  </Grid>
                </Grid>
              }
            </CardActionArea>
          </Grid> */}
          <Grid>
            <Typography variant="h6">Settings</Typography>
          </Grid>
          <Grid>
            <PickTheme {...props} />
          </Grid>
        </Grid>
      </Grid>
      <Grid size={{ sm: 12, md: 8 }}>
        <Grid container direction="column" spacing={2}>
          <Grid>
            <Typography variant="h6">Group</Typography>
          </Grid>
          <Grid>
            <ManageGroups {...props} />
          </Grid>
        </Grid>
      </Grid>
      <Grid size={12}>
        <Button sx={classes.red} onClick={handleSubmit}>Submit</Button>
      </Grid>
    </Grid>
  </>
}

export default Profile;
