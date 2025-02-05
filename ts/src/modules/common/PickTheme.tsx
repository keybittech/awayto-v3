import React from 'react';

import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

import { useStyles, useTheme, PaletteMode } from 'awayto/hooks';
declare global {
  interface IComponent {
    showTitle?: boolean;
  }
}

export function PickTheme(props: IComponent): React.JSX.Element {
  const { showTitle } = props;

  const classes = useStyles();

  const { setTheme } = useTheme();

  const edit = (e: React.SyntheticEvent) => {
    localStorage.setItem('site_theme', e.currentTarget.id);
    setTheme({ variant: e.currentTarget.id as PaletteMode });
  };

  return <>
    <Grid container alignItems="center">
      {showTitle ? <Grid><Typography>Theme</Typography></Grid> : <></>}
      <Grid onClick={edit} id="dark"><Box bgcolor="gray" sx={classes.colorBox} /></Grid>
      <Grid onClick={edit} id="light"><Box bgcolor="white" sx={classes.colorBox} /></Grid>
      {/* <Grid onClick={edit} id="blue"><Box bgcolor="deepskyblue" sx={classes.colorBox} /></Grid> */}
    </Grid>
  </>
}

export default PickTheme;
