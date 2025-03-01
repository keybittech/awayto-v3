import React, { useMemo } from 'react';

import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardActionArea from '@mui/material/CardActionArea';
import Dialog from '@mui/material/Dialog';

import { useAppSelector, useUtil, getUtilRegisteredAction, targets } from 'awayto/hooks';
import { CardHeader } from '@mui/material';

export function ConfirmAction(): React.JSX.Element {

  const { closeConfirm } = useUtil();
  const util = useAppSelector(state => state.util);

  return useMemo(() => <Dialog open={!!util.isConfirming} fullWidth={true} maxWidth="sm">
    <Card>
      <CardHeader title="Confirm Action" subheader={`Action: ${util.confirmEffect}`} />
      <Grid container sx={{ minHeight: '25vh' }}>
        <Grid size={{ xs: util.confirmSideEffect ? 6 : 12 }}>
          <CardActionArea
            {...targets(`confirmation approval`, `confirm approval for ${util.confirmSideEffect ? util.confirmSideEffect.approvalEffect : 'general confirmation'}`)}
            sx={{ height: '100%', padding: '50px' }}
            onClick={() => {
              async function go() {
                await getUtilRegisteredAction(util.confirmActionId)(true);
                closeConfirm({});
              }
              void go();
            }}>
            <Grid container textAlign="center" justifyContent="center">
              <Grid>
                {util.confirmSideEffect && <Typography variant="button" fontSize={16}>{util.confirmSideEffect?.approvalAction}</Typography>}
              </Grid>
              <Grid>
                <Typography variant="caption">{util.confirmSideEffect?.approvalEffect ? 'Click here to: ' + util.confirmSideEffect.approvalEffect : 'Click here to confirm.'}</Typography>
              </Grid>
            </Grid>
          </CardActionArea>
        </Grid>
        {util.confirmSideEffect && <Grid size={6}>

          <CardActionArea
            {...targets(`confirmation rejection`, `confirm rejection for ${util.confirmSideEffect.rejectionEffect}`)}
            sx={{ height: '100%', padding: '50px' }}
            onClick={() => {
              async function go() {
                await getUtilRegisteredAction(util.confirmActionId)(false);
                closeConfirm({});
              }
              void go();
            }}>
            <Grid container textAlign="center" justifyContent="center">
              <Typography variant="button" fontSize={16}>{util.confirmSideEffect.rejectionAction}</Typography>
              <Typography variant="caption">Click here to: {util.confirmSideEffect.rejectionEffect}</Typography>
            </Grid>
          </CardActionArea>
        </Grid>}
      </Grid>
      <CardActions>
        <Button
          {...targets(`cancel confirmation`, `exit the confirmation screen`)}
          onClick={() => { closeConfirm({}); }}
        >Cancel</Button>
      </CardActions>
    </Card>
  </Dialog >, [util]);
}

export default ConfirmAction;
