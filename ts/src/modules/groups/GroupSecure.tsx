import React, { useEffect, useState, ReactNode } from 'react';
import { useSecure, SiteRoles } from 'awayto/hooks';

declare global {
  interface IComponent {
    contentGroupRoles?: SiteRoles[];
    children?: ReactNode;
  }
}

export function GroupSecure({ contentGroupRoles = [SiteRoles.APP_GROUP_ADMIN], children }: IComponent): React.JSX.Element {

  const hasRole = useSecure();

  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    setIsValid(hasRole(contentGroupRoles));
  }, [hasRole]);

  return <> {isValid ? children : <></>} </>
}

export default GroupSecure;
