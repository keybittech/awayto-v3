import React, { useEffect, useState } from 'react';
import { useSecure, SiteRoles } from 'awayto/hooks';

interface GroupSecureProps extends IComponent {
  contentGroupRoles: SiteRoles[];
}

export function GroupSecure({ contentGroupRoles = [SiteRoles.APP_GROUP_ADMIN], children }: GroupSecureProps): React.JSX.Element {

  const hasRole = useSecure();

  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    setIsValid(hasRole(contentGroupRoles));
  }, [hasRole]);

  return <> {isValid ? children : <></>} </>
}

export default GroupSecure;
