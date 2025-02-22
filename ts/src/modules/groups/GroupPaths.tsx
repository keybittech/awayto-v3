import React from 'react';
import { Route, Routes } from 'react-router';
import ManageGroupHome from './ManageGroupHome';

export function GroupPaths(props: IComponent): React.JSX.Element {
  return <Routes>
    <Route path="manage" element={<ManageGroupHome {...props} />} />
    <Route path="manage/:component" element={<ManageGroupHome {...props} />} />
  </Routes>
}

export default GroupPaths;
