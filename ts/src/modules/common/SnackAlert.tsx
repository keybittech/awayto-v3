import React, { useEffect } from 'react';
import Snackbar from '@mui/material/Snackbar';
import Box from '@mui/material/Box';
import MuiAlert, { AlertProps } from '@mui/material/Alert';
import { useUtil, useAppSelector } from 'awayto/hooks';

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref,
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

export function SnackAlert(): React.JSX.Element {

  const { setSnack } = useUtil();
  const { snackOn, snackType, snackRequestId } = useAppSelector(state => state.util);

  const hideSnack = (): void => {
    setSnack({ snackOn: '', snackRequestId: '' });
  }

  return !snackOn ? <></> : <Snackbar
    sx={{
      zIndex: 10000,
      '.MuiSvgIcon-root': {
        color: 'black'
      },
      whiteSpace: 'pre-wrap',
      position: 'fixed'
    }}
    anchorOrigin={{
      vertical: 'top',
      horizontal: 'center'
    }}
    open={!!snackOn}
    // autoHideDuration={15000}
    onClose={hideSnack}
  >
    <Alert onClose={hideSnack} severity={snackType || "info"}>
      <Box>{snackOn}</Box>
      <Box><sub>{snackRequestId}</sub></Box>
    </Alert>
  </Snackbar>
}

export default SnackAlert;
