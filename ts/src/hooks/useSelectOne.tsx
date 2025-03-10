import React, { useEffect, useState } from 'react';

import TextField, { TextFieldProps } from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';

import { ILookup } from './api';
import { targets } from './util';

export type UseSelectOneResponse<T> = {
  item?: T;
  setId: (id: string) => void;
  comp: (props: IComponent & TextFieldProps) => React.JSX.Element;
};

export function useSelectOne<T extends Partial<ILookup>>(label: string, { data: items }: { data?: T[] }): UseSelectOneResponse<T> {
  const [itemId, setItemId] = useState(Array.isArray(items) && items.length ? items[0].id : '');

  useEffect(() => {
    if (Array.isArray(items) && items.length) {
      const currentItem = items.find(it => it.id === itemId);
      if (!currentItem) {
        const firstItem = items[0];
        if (firstItem) {
          setItemId(firstItem.id);
        }
      }
    } else {
      setItemId('');
    }
  }, [items, itemId]);

  const handleMenuItemClick = (id: string) => {
    if (id !== itemId) {
      setItemId(id);
    }
  };

  return {
    item: (items ? items?.find(it => it.id === itemId) : {}) as T,
    setId: setItemId,
    comp: ({ children, ...props }) => items?.length ? <TextField
      {...targets(`select one selection ${label}`, label, `make a selection for ${label}`)}
      select
      fullWidth
      value={itemId}
      onChange={e => {
        handleMenuItemClick(e.target.value);
      }}
      {...props}
    >
      {children ? children : items.map((it, i) =>
        <MenuItem key={i} value={it.id}>{it.name}</MenuItem>
      )}
    </TextField> : <></>
  }
}
