import React, { useEffect, useState } from 'react';
import { useSecure, SiteRoles } from 'awayto/hooks';

interface GroupSecureProps extends IComponent {
  contentGroupRoles: SiteRoles[];
}

export function GroupSecure({ contentGroupRoles = [SiteRoles.APP_GROUP_ADMIN], children }: GroupSecureProps): React.JSX.Element {

  const secure = useSecure();

  const [isValid, setIsValid] = useState(false);

  useEffect(() => {
    setIsValid(secure(contentGroupRoles));
  }, [secure]);

  return <> {isValid ? children : <></>} </>
}

export default GroupSecure;
