import React, { useContext } from 'react';

import Divider from '@mui/material/Divider';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

import ApprovalIcon from '@mui/icons-material/Approval';
import DoNotDisturbIcon from '@mui/icons-material/DoNotDisturb';
import CheckBoxOutlineBlank from '@mui/icons-material/CheckBoxOutlineBlank';
import CheckBox from '@mui/icons-material/CheckBox';

import { bookingFormat, targets } from 'awayto/hooks';

import PendingQuotesContext, { PendingQuotesContextType } from './PendingQuotesContext';

interface PendingQuotesMenuProps extends IComponent {
  pendingQuotesAnchorEl: null | HTMLElement;
  pendingQuotesMenuId: string;
  isPendingQuotesOpen: boolean;
  handleMenuClose: () => void;
}

export function PendingQuotesMenu({ handleMenuClose, pendingQuotesAnchorEl, pendingQuotesMenuId, isPendingQuotesOpen }: PendingQuotesMenuProps): React.JSX.Element {

  const {
    pendingQuotes,
    selectedPendingQuotes,
    handleSelectPendingQuote,
    handleSelectPendingQuoteAll,
    approvePendingQuotes,
    denyPendingQuotes
  } = useContext(PendingQuotesContext) as PendingQuotesContextType;

  const hasSelections = Boolean(selectedPendingQuotes.length);

  return <Menu
    anchorEl={pendingQuotesAnchorEl}
    anchorOrigin={{
      vertical: 'bottom',
      horizontal: 'right',
    }}
    id={pendingQuotesMenuId}
    transformOrigin={{
      vertical: 'top',
      horizontal: 'right',
    }}
    open={!!pendingQuotes.length && isPendingQuotesOpen}
    onClose={handleMenuClose}
    MenuListProps={{
      'aria-labelledby': 'topbar-pending-toggle'
    }}
  >
    <MenuItem
      {...targets(`pending requests menu select all`, `select all pending requests in the list`)}
      onClick={handleSelectPendingQuoteAll}
    >
      <ListItemIcon>
        {selectedPendingQuotes.length === pendingQuotes.length && pendingQuotes.length !== 0 ?
          <CheckBox color="success" /> :
          <CheckBoxOutlineBlank />
        }
      </ListItemIcon>
      <ListItemText primary="Select All" />
    </MenuItem>

    {hasSelections && <MenuItem
      {...targets(`pending requests menu approve`, `approve selected pending requests`)}
      onClick={approvePendingQuotes}
    >
      <ListItemIcon><ApprovalIcon /></ListItemIcon>
      <ListItemText primary="Approve" />
    </MenuItem>}

    {hasSelections && <MenuItem
      {...targets(`pending requests menu deny`, `deny selected pending requests`)}
      onClick={denyPendingQuotes}
    >
      <ListItemIcon><DoNotDisturbIcon /></ListItemIcon>
      <ListItemText primary="Deny" />
    </MenuItem>}

    <Divider />

    {pendingQuotes.map((pq, i) => {
      return <MenuItem
        key={`pending_quotes_pqs_${i}`}
        {...targets(`pending requests menu select ${i}`, `select a single pending request from the list`)}
        onClick={() => handleSelectPendingQuote(pq.id)}
      >
        <ListItemIcon>
          {selectedPendingQuotes.indexOf(pq.id) !== -1 ?
            <CheckBox color="success" /> :
            <CheckBoxOutlineBlank />
          }
        </ListItemIcon>
        <ListItemText
          primary={`${bookingFormat(pq.slotDate, pq.startTime)}`}
          secondary={`${pq.serviceName} ${pq.serviceTierName}`}
        />
      </MenuItem>
    })}
  </Menu>
}

export default PendingQuotesMenu;
