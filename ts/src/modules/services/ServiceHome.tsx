import React, { Suspense, useEffect, useMemo, useState } from 'react';

import TextField from '@mui/material/TextField';
import Slider from '@mui/material/Slider';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import MenuItem from '@mui/material/MenuItem';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActionArea from '@mui/material/CardActionArea';
import CardHeader from '@mui/material/CardHeader';
import Grid from '@mui/material/Grid';
import Chip from '@mui/material/Chip';

import { useComponents, IService, useStyles, useUtil, useSuggestions } from 'awayto/hooks';

export function ServiceHome(props: IProps): React.JSX.Element {

  const { ManageServiceModal } = useComponents();

  const [service, setService] = useState({} as IService);


  return <Suspense>
    <ManageServiceModal editService={service} />
  </Suspense>;
}

export const roles = [];

export default ServiceHome;
