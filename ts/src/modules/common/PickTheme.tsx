import React from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

import { useStyles, PaletteMode } from 'awayto/hooks';
import { useColorScheme } from '@mui/material';

declare global {
  interface IComponent {
    showTitle?: boolean;
  }
}

export function PickTheme(props: IComponent): React.JSX.Element {
  const { showTitle } = props;

  const classes = useStyles();

  const { setMode } = useColorScheme();

  const editMode = (e: React.SyntheticEvent) => {
    setMode(e.currentTarget.id as PaletteMode);
  };

  return <>
    <Grid container alignItems="center">
      {showTitle ? <Grid><Typography>Theme</Typography></Grid> : <></>}
      <Grid onClick={editMode} id="dark"><Box bgcolor="gray" sx={classes.colorBox} /></Grid>
      <Grid onClick={editMode} id="light"><Box bgcolor="white" sx={classes.colorBox} /></Grid>
    </Grid>
  </>
}

export default PickTheme;
