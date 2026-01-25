import React from 'react';
import { Navigate, Route, Routes } from 'react-router';
import ManageGroupHome from './ManageGroupHome';
import FormReport from '../forms/FormReport';

export function GroupPaths(props: IComponent): React.JSX.Element {
  return <Routes>
    <Route path="manage" element={<ManageGroupHome {...props} />} />
    <Route path="manage/:component" element={<ManageGroupHome {...props} />} />
    <Route path="manage/forms/:formId/report" element={<FormReport {...props} />} />
    <Route path="*" element={<Navigate replace to="/" />} />
  </Routes>
}

export default GroupPaths;
