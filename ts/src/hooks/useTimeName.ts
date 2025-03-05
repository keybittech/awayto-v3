import { siteApi } from './api';

export function useTimeName(id?: string): string | undefined {
  const { data: lookups } = siteApi.useLookupServiceGetLookupsQuery();
  if (lookups) {
    const tu = lookups.timeUnits.find(tu => tu.id == id);
    if (tu) {
      return tu.name;
    }
  }
  return undefined;
}
