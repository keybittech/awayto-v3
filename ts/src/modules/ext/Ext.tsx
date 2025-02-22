import React from 'react';
import { Route, Routes } from 'react-router-dom';
import Kiosk from '../kiosk/Kiosk';

export function Ext(): React.JSX.Element {

  return <>
    <Routes>
      <Route path="/kiosk/:groupName/:scheduleName?" element={
        <Kiosk />
      } />
    </Routes>
  </>
}

export default Ext;
