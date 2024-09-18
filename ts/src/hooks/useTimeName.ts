import { useMemo } from 'react';
import { siteApi } from './api';

export function useTimeName(id?: string): string {
  const { data: lookups } = siteApi.useLookupServiceGetLookupsQuery();
  const timeName = useMemo(() => {
    if (lookups) {
      const tu = lookups.timeUnits.find(tu => tu.id == id);
      if (tu) {
        return tu.name || '';
      }
    }
    return '';
  }, [id, lookups])

  return timeName;
}
