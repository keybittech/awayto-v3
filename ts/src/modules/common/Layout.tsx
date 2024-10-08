import React, { useEffect, Suspense } from 'react';
import { Route, Outlet, Routes } from 'react-router-dom';

import Grid from '@mui/material/Unstable_Grid2';


import { useComponents } from 'awayto/hooks';
import Sidebar from './Sidebar';
import Topbar from './Topbar';

const Layout = (props: IComponent): React.JSX.Element => {

  const {
    Home,
    Exchange,
    ExchangeSummary,
    ExchangeProvider,
    TopLevelProviders,
    Profile,
    GroupPaths,
    ServiceHome,
    ScheduleHome,
    RequestQuote
  } = useComponents();

  useEffect(() => {
    window.INT_SITE_LOAD = true;
  }, []);

  return <>
    <TopLevelProviders>
      <Routes>
        <Route element={
          <Grid container direction="row">
            {/* <Grid width={175} sx={{ bgcolor: 'primary.dark', position: 'fixed', minWidth: '175px', display: { xs: 'none', md: 'flex' } }}>
              <Sidebar />
            </Grid> */}
            <Grid xs={12} container direction="column" sx={{ marginLeft: { xs: 0, md: true ? 0 : '175px' } }}>
              <Grid px={1} sx={{ bgcolor: 'primary.dark' }}>
                <Topbar {...props} />
              </Grid>
              <Grid p={2} sx={{ width: '100%', minHeight: 'calc(100vh - 75px)' }}>
                <Suspense>
                  <Outlet />
                </Suspense>
              </Grid>
            </Grid>
          </Grid>
        }>
          <Route path="/" element={<Home {...props} />} />
          <Route path="/profile" element={<Profile {...props} />} />
          <Route path="/service" element={<ServiceHome {...props} />} />
          <Route path="/quote/request" element={<RequestQuote {...props} />} />
          <Route path="/schedule" element={<ScheduleHome {...props} />} />
          <Route path="/group/*" element={<GroupPaths {...props} />} />
          <Route path="/exchange/:summaryId/summary" element={<ExchangeSummary {...props} />} />
        </Route>
        <Route element={
          <Grid xs={12} container direction="column">
            <Grid xs={12} px={1} sx={{ bgcolor: 'primary.dark' }}>
              <Topbar forceSiteMenu={true} {...props} />
            </Grid>
            <Grid sx={{ display: 'flex', height: 'calc(100vh - 60px)', width: '100%' }}>
              <Suspense>
                <ExchangeProvider>
                  <Outlet />
                </ExchangeProvider>
              </Suspense>
            </Grid>
          </Grid>
        }>
          <Route path="/exchange/:exchangeId" element={<Exchange {...props} />} />
        </Route>
      </Routes>
    </TopLevelProviders>
  </>
}

export default Layout;
